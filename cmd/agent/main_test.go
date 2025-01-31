package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_updateMetrics(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		wantType string
	}{
		{name: "Alloc", metric: "Alloc", wantType: "gauge"},
		{name: "RandomValue", metric: "RandomValue", wantType: "gauge"},
		{name: "PollCount", metric: "PollCount", wantType: "counter"},
	}

	metrics := updateMetrics()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			found := false
			for _, metric := range metrics {
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
