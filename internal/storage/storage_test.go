package storage

import (
	"context"
	"fmt"
	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStorage_Save(t *testing.T) {
	cfg := config.ConfigureServer()
	store := New(context.Background(), cfg)

	gauge := 146.33
	delta := int64(10)
	var batch []model.Metrics
	batch = append(batch, model.Metrics{ID: "MAlloc", MType: "gauge", Value: &gauge})
	batch = append(batch, model.Metrics{ID: "Counter3", MType: "counter", Delta: &delta})

	tests := []struct {
		name          string
		mtype         string
		mname         string
		mvalue        string
		expectedValue interface{}
	}{
		{name: "gauge", mtype: "gauge", mname: "Alloc", mvalue: "123", expectedValue: float64(123)},
		{name: "counter", mtype: "counter", mname: "Count", mvalue: "10", expectedValue: int64(10)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.Save(tt.mtype, tt.mname, tt.mvalue)
			value, _ := store.Get(tt.mtype, tt.mname)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

func TestStorage_SaveMetric(t *testing.T) {
	cfg := config.ServerConfig{}
	store := New(context.Background(), cfg)

	gauge := 146.33
	delta := int64(10)

	tests := []struct {
		name   string
		metric model.Metrics
	}{
		{
			name:   "gauge",
			metric: model.Metrics{ID: "MAlloc", MType: "gauge", Value: &gauge},
		},
		{
			name:   "counter",
			metric: model.Metrics{ID: "Counter2", MType: "counter", Delta: &delta},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.SaveMetric(tt.metric)
			metric := model.Metrics{ID: tt.metric.ID, MType: tt.metric.MType}
			store.GetMetric(&metric)
			assert.Equal(t, tt.metric, metric)
		})
	}
}

func TestStorage_SetAllMetric(t *testing.T) {
	cfg := config.ServerConfig{}
	store := New(context.Background(), cfg)

	gauge := 146.33
	delta := int64(10)
	var batch []model.Metrics
	batch = append(batch, model.Metrics{ID: "MAlloc", MType: "gauge", Value: &gauge})
	batch = append(batch, model.Metrics{ID: "Counter3", MType: "counter", Delta: &delta})

	tests := []struct {
		name    string
		metrics []model.Metrics
	}{
		{
			name:    "batch",
			metrics: batch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.SetAllMetrics(tt.metrics)

			res, _ := store.GetAllMetrics()
			fmt.Println(res)
			assert.Equal(t, tt.metrics, res)
		})
	}
}

func TestStorage_Backup(t *testing.T) {
	cfg := config.ServerConfig{StoragePath: "/tmp/test_metrics_backup.json"}
	store := New(context.Background(), cfg)

	gauge := 146.33
	delta := int64(10)
	var batch []model.Metrics
	batch = append(batch, model.Metrics{ID: "MAlloc", MType: "gauge", Value: &gauge})
	batch = append(batch, model.Metrics{ID: "Counter1", MType: "counter", Delta: &delta})

	t.Run("backup", func(t *testing.T) {
		store.SetAllMetrics(batch)
		err := store.Backup()
		if err != nil {
			t.Fatal(err)
		}
	})
}
