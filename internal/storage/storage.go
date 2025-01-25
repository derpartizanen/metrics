package storage

import (
	"errors"
	"strconv"

	"github.com/derpartizanen/metrics/internal/model"
)

const (
	TypeCounter = "counter"
	TypeGauge   = "gauge"
)

var (
	ErrInvalidGaugeMetricValue   = errors.New("invalid gauge metric value")
	ErrInvalidCounterMetricValue = errors.New("invalid counter metric value")
	ErrInvalidMetricType         = errors.New("invalid metric type")
)

type Storage struct {
	repository Repository
}

type Repository interface {
	UpdateCounterMetric(string, int64) error
	UpdateGaugeMetric(string, float64) error
	GetGaugeMetric(string) (float64, error)
	GetCounterMetric(string) (int64, error)
	GetAllMetrics() (map[string]float64, map[string]int64, error)
}

func New(r Repository) *Storage {
	return &Storage{repository: r}
}

func (s *Storage) Save(metricType string, metricName string, value string) error {
	if metricType == TypeCounter {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrInvalidCounterMetricValue
		}

		return s.repository.UpdateCounterMetric(metricName, intValue)
	}

	if metricType == TypeGauge {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrInvalidGaugeMetricValue
		}

		return s.repository.UpdateGaugeMetric(metricName, floatValue)
	}

	return ErrInvalidMetricType
}

func (s *Storage) SaveMetric(metric model.Metrics) error {
	if metric.MType == TypeCounter {
		return s.repository.UpdateCounterMetric(metric.ID, *metric.Delta)
	}

	if metric.MType == TypeGauge {
		return s.repository.UpdateGaugeMetric(metric.ID, *metric.Value)
	}

	return ErrInvalidMetricType
}

func (s *Storage) Get(metricType string, metricName string) (interface{}, error) {
	if metricType == TypeGauge {
		value, err := s.repository.GetGaugeMetric(metricName)

		return value, err
	}

	if metricType == TypeCounter {
		value, err := s.repository.GetCounterMetric(metricName)

		return value, err
	}

	return nil, ErrInvalidMetricType
}

func (s *Storage) GetMetric(metric *model.Metrics) error {
	if metric.MType == TypeGauge {
		value, err := s.repository.GetGaugeMetric(metric.ID)
		if err != nil {
			return err
		}
		metric.Value = &value
	}

	if metric.MType == TypeCounter {
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
