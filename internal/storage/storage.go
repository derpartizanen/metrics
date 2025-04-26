package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/interfaces"
	"github.com/derpartizanen/metrics/internal/repository/memstorage"
	"github.com/derpartizanen/metrics/internal/repository/postgres"
	"os"
	"strconv"

	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/model"
	"go.uber.org/zap"
)

var (
	ErrInvalidGaugeMetricValue   = errors.New("invalid gauge metric value")
	ErrInvalidCounterMetricValue = errors.New("invalid counter metric value")
	ErrInvalidMetricType         = errors.New("invalid metric type")
)

type Storage struct {
	repository interfaces.Repository
	settings   Settings
}

type Settings struct {
	StoragePath   string
	StoreInterval int64
}

func New(ctx context.Context, cfg config.ServerConfig) *Storage {
	settings := Settings{
		StoragePath:   cfg.StoragePath,
		StoreInterval: cfg.StoreInterval,
	}

	if cfg.DatabaseDSN != "" {
		repo, err := postgres.New(ctx, cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("Init database storage error", zap.Error(err))
		}

		return &Storage{repository: repo, settings: settings}
	}

	repo := memstorage.New()
	storage := &Storage{repository: repo, settings: settings}
	if cfg.Restore {
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

	f, err := os.Create(s.settings.StoragePath)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "    ")

	metrics, err := s.GetAllMetrics()
	if err != nil {
		return err
	}

	if err := encoder.Encode(metrics); err != nil {
		return err
	}

	return writer.Flush()
}

func (s *Storage) Save(metricType string, metricName string, value string) error {
	if metricType == model.MetricTypeCounter {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrInvalidCounterMetricValue
		}

		return s.repository.UpdateCounterMetric(metricName, intValue)
	}

	if metricType == model.MetricTypeGauge {
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

	if metric.MType == model.MetricTypeCounter {
		if metric.Delta == nil {
			return ErrInvalidCounterMetricValue
		}
		err = s.repository.UpdateCounterMetric(metric.ID, *metric.Delta)
	} else if metric.MType == model.MetricTypeGauge {
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
	if metricType == model.MetricTypeGauge {
		value, err := s.repository.GetGaugeMetric(metricName)

		return value, err
	}

	if metricType == model.MetricTypeCounter {
		value, err := s.repository.GetCounterMetric(metricName)

		return value, err
	}

	return nil, ErrInvalidMetricType
}

func (s *Storage) GetMetric(metric *model.Metrics) error {
	if metric.MType == model.MetricTypeGauge {
		value, err := s.repository.GetGaugeMetric(metric.ID)
		if err != nil {
			return err
		}
		metric.Value = &value
	}

	if metric.MType == model.MetricTypeCounter {
		value, err := s.repository.GetCounterMetric(metric.ID)
		if err != nil {
			return err
		}
		metric.Delta = &value
	}

	return nil
}

func (s *Storage) GetAllMetrics() ([]model.Metrics, error) {
	metrics, err := s.repository.GetAllMetrics()
	if err != nil {
		logger.Log.Error("Get metrics error")
	}

	return metrics, nil
}

func (s *Storage) SetAllMetrics(metrics []model.Metrics) error {
	return s.repository.SetAllMetrics(metrics)
}

func (s *Storage) Ping() error {
	return s.repository.Ping()
}
