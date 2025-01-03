package main

import (
	"net/http"
	"strconv"
)

func main() {
	metricHandler := &MetricHandler{
		storage: &MemStorage{
			gauge:   make(map[string]float64),
			counter: make(map[string]int64),
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{metricType}/{metricName}/{metricValue}", metricHandler.updateMetric)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

// Абстрактное хранилище
type MetricUpdaterGetter interface {
	UpdateGaugeMetric(name string, value float64)
	UpdateCounterMetric(name string, value int64)
}

// Обработчик
type MetricHandler struct {
	storage MetricUpdaterGetter
}

func (h *MetricHandler) updateMetric(res http.ResponseWriter, req *http.Request) {
	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")
	metricValue := req.PathValue("metricValue")

	if metricType != "gauge" && metricType != "counter" {
		http.Error(res, "Unknown metric type", http.StatusBadRequest)
		return
	}

	if metricType == "gauge" {
		gaugeValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(res, "Incorrect metric value for type gauge", http.StatusBadRequest)
			return
		}

		h.storage.UpdateGaugeMetric(metricName, gaugeValue)
	}

	if metricType == "counter" {
		counterValue, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(res, "Incorrect metric value for type counter", http.StatusBadRequest)
			return
		}

		h.storage.UpdateCounterMetric(metricName, counterValue)
	}
}

// Хранилище
type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func (s *MemStorage) UpdateGaugeMetric(name string, value float64) {
	s.gauge[name] = value
}

func (s *MemStorage) UpdateCounterMetric(name string, value int64) {
	s.counter[name] += value
}
