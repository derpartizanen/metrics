package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/derpartizanen/metrics/internal/model"
)

func TestAgent_CollectMemStatsMetrics(t *testing.T) {
	metricsAgent := Agent{
		Metrics: make([]model.Metrics, 0),
	}

	tests := []struct {
		name     string
		metric   string
		wantType string
	}{
		{name: "Alloc", metric: "Alloc", wantType: "gauge"},
		{name: "RandomValue", metric: "RandomValue", wantType: "gauge"},
		{name: "PollCount", metric: "PollCount", wantType: "counter"},
	}

	metricsAgent.CollectMemStatsMetrics()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			found := false
			for _, metric := range metricsAgent.Metrics {
				if metric.ID == test.metric {
					found = true
					assert.Equal(t, metric.MType, test.wantType)
					break
				}
			}
			if !found {
				t.Errorf("metric %s not found", test.metric)
			}
		})
	}
}

func TestAgent_CollectPsutilMetrics(t *testing.T) {
	metricsAgent := Agent{
		Metrics: make([]model.Metrics, 0),
	}

	tests := []struct {
		name     string
		metric   string
		wantType string
	}{
		{name: "TotalMemory", metric: "TotalMemory", wantType: "gauge"},
		{name: "FreeMemory", metric: "FreeMemory", wantType: "gauge"},
		{name: "CPUutilization1", metric: "CPUutilization1", wantType: "gauge"},
	}

	metricsAgent.CollectPsutilMetrics()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			found := false
			for _, metric := range metricsAgent.Metrics {
				if metric.ID == test.metric {
					found = true
					assert.Equal(t, metric.MType, test.wantType)
					break
				}
			}
			if !found {
				t.Errorf("metric %s not found", test.metric)
			}
		})
	}
}
