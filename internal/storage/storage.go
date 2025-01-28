package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/model"
	"go.uber.org/zap"
)

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

var (
	ErrInvalidGaugeMetricValue   = errors.New("invalid gauge metric value")
	ErrInvalidCounterMetricValue = errors.New("invalid counter metric value")
	ErrInvalidMetricType         = errors.New("invalid metric type")
)

type Storage struct {
	repository Repository
	settings   Settings
}

type Settings struct {
	StoragePath   string
	StoreInterval int64
	Restore       bool
}

type Repository interface {
	UpdateCounterMetric(string, int64) error
	UpdateGaugeMetric(string, float64) error
	GetGaugeMetric(string) (float64, error)
	GetCounterMetric(string) (int64, error)
	GetAllMetrics() (map[string]float64, map[string]int64, error)
	SetAllMetrics(metrics []model.Metrics) error
}

func New(r Repository, s Settings) *Storage {
	storage := &Storage{repository: r, settings: s}

	if storage.settings.Restore {
		err := storage.Restore()
		if err != nil {
			logger.Log.Error("Restore failed", zap.Error(err))
		}
	}

	return storage
}

func (s *Storage) Restore() error {
	logger.Log.Info("Restoring metrics from backup file")

	data, err := os.ReadFile(s.settings.StoragePath)
	if err != nil {
		return err
	}

	metrics := make([]model.Metrics, 0)
	if err := json.Unmarshal(data, &metrics); err != nil {
		return err
	}

	logger.Log.Info(fmt.Sprintf("Loaded %d metrics", len(metrics)))

	return s.SetAllMetrics(metrics)
}

func (s *Storage) Backup() error {
	logger.Log.Debug("Backing up metrics to file")
	metrics, err := s.GetAllMetrics()
	metricsJson, err := json.MarshalIndent(metrics, "", "   ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.settings.StoragePath, metricsJson, 0666)
}

func (s *Storage) Save(metricType string, metricName string, value string) error {
	if metricType == MetricTypeCounter {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrInvalidCounterMetricValue
		}

		return s.repository.UpdateCounterMetric(metricName, intValue)
	}

	if metricType == MetricTypeGauge {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrInvalidGaugeMetricValue
		}

		return s.repository.UpdateGaugeMetric(metricName, floatValue)
	}

	return ErrInvalidMetricType
}

func (s *Storage) SaveMetric(metric model.Metrics) error {
	var err error

	if metric.MType == MetricTypeCounter {
		if metric.Delta == nil {
			return ErrInvalidCounterMetricValue
		}
		err = s.repository.UpdateCounterMetric(metric.ID, *metric.Delta)
	} else if metric.MType == MetricTypeGauge {
		err = s.repository.UpdateGaugeMetric(metric.ID, *metric.Value)
	} else {
		err = ErrInvalidMetricType
	}

	if err != nil {
		return err
	}

	if s.settings.StoreInterval == 0 {
		logger.Log.Debug("Sync metrics save")
		err = s.Backup()
		if err != nil {
			logger.Log.Error("Sync save failed", zap.Error(err))
		}
	}

	return nil
}

func (s *Storage) Get(metricType string, metricName string) (interface{}, error) {
	if metricType == MetricTypeGauge {
		value, err := s.repository.GetGaugeMetric(metricName)

		return value, err
	}

	if metricType == MetricTypeCounter {
		value, err := s.repository.GetCounterMetric(metricName)

		return value, err
	}

	return nil, ErrInvalidMetricType
}

func (s *Storage) GetMetric(metric *model.Metrics) error {
	if metric.MType == MetricTypeGauge {
		value, err := s.repository.GetGaugeMetric(metric.ID)
		if err != nil {
			return err
		}
		metric.Value = &value
	}

	if metric.MType == MetricTypeCounter {
		value, err := s.repository.GetCounterMetric(metric.ID)
		if err != nil {
			return err
		}
		metric.Delta = &value
	}

	return nil
}

func (s *Storage) GetAll() ([]model.Metric, error) {
	gauges, counters, _ := s.repository.GetAllMetrics()

	var metrics []model.Metric
	for name, value := range gauges {
		metrics = append(metrics, model.Metric{Name: name, Type: "gauge", Value: value})
	}
	for name, value := range counters {
		metrics = append(metrics, model.Metric{Name: name, Type: "counter", Value: value})
	}

	return metrics, nil
}

func (s *Storage) GetAllMetrics() ([]model.Metrics, error) {
	gauges, counters, _ := s.repository.GetAllMetrics()

	var metrics []model.Metrics
	for name, value := range gauges {
		metrics = append(metrics, model.Metrics{ID: name, MType: "gauge", Value: &value})
	}
	for name, value := range counters {
		metrics = append(metrics, model.Metrics{ID: name, MType: "counter", Delta: &value})
	}

	return metrics, nil
}

func (s *Storage) SetAllMetrics(metrics []model.Metrics) error {
	return s.repository.SetAllMetrics(metrics)
}
