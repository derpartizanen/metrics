package interfaces

import "github.com/derpartizanen/metrics/internal/model"

type Repository interface {
	UpdateCounterMetric(string, int64) error
	UpdateGaugeMetric(string, float64) error
	GetGaugeMetric(string) (float64, error)
	GetCounterMetric(string) (int64, error)
	GetAllMetrics() ([]model.Metrics, error)
	SetAllMetrics(metrics []model.Metrics) error
	Ping() error
}
