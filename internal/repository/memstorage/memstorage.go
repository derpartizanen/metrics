package memstorage

import (
	"errors"

	"github.com/derpartizanen/metrics/internal/model"
)

var (
	ErrNotFound = errors.New("value not found")
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func New() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (s *MemStorage) UpdateGaugeMetric(name string, value float64) error {
	s.gauge[name] = value

	return nil
}

func (s *MemStorage) UpdateCounterMetric(name string, value int64) error {
	s.counter[name] += value

	return nil
}

func (s *MemStorage) GetGaugeMetric(metricName string) (float64, error) {
	value, ok := s.gauge[metricName]
	if ok {
		return value, nil
	}

	return 0, ErrNotFound
}

func (s *MemStorage) GetCounterMetric(metricName string) (int64, error) {
	value, ok := s.counter[metricName]
	if ok {
		return value, nil
	}

	return 0, ErrNotFound
}

func (s *MemStorage) GetAllMetrics() ([]model.Metrics, error) {
	var metrics []model.Metrics
	for name, value := range s.gauge {
		metrics = append(metrics, model.Metrics{ID: name, MType: "gauge", Value: &value})
	}
	for name, value := range s.counter {
		metrics = append(metrics, model.Metrics{ID: name, MType: "counter", Delta: &value})
	}

	return metrics, nil
}

func (s *MemStorage) SetAllMetrics(metrics []model.Metrics) error {
	for _, metric := range metrics {
		if metric.MType == model.MetricTypeCounter {
			err := s.UpdateCounterMetric(metric.ID, *metric.Delta)
			if err != nil {
				return err
			}
		}
		if metric.MType == model.MetricTypeGauge {
			err := s.UpdateGaugeMetric(metric.ID, *metric.Value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *MemStorage) Ping() error {
	return nil
}
