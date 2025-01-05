package storage

import (
	"errors"
	"strconv"
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
