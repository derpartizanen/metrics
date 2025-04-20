package agent

import (
	"github.com/derpartizanen/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_collectMetrics(t *testing.T) {
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
