package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/derpartizanen/metrics/internal/model"
	"github.com/derpartizanen/metrics/internal/storage"
)

type Handler struct {
	storage *storage.Storage
}

func NewHandler(storage *storage.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

func (h *Handler) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")
	metricValue := req.PathValue("metricValue")

	err := h.storage.Save(metricType, metricName, metricValue)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
}

func (h *Handler) GetHandler(res http.ResponseWriter, req *http.Request) {
	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")

	var result string

	value, err := h.storage.Get(metricType, metricName)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidMetricType) {
			http.Error(res, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(res, err.Error(), http.StatusNotFound)
		}
	}

	if metricType == storage.TypeCounter {
		result = fmt.Sprintf("%d", value)
	} else {
		result = fmt.Sprintf("%g", value)
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, result)
}

func (h *Handler) GetAllHandler(res http.ResponseWriter, req *http.Request) {
	var result string
	metrics, _ := h.storage.GetAll()

	for _, metric := range metrics {
		if metric.Type == storage.TypeCounter {
			result += fmt.Sprintf("%s: %d\n", metric.Name, metric.Value)
		}
		if metric.Type == storage.TypeGauge {
			result += fmt.Sprintf("%s: %f\n", metric.Name, metric.Value)
		}

	}
	res.Header().Set("Content-Type", "text/html")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, result)
}

func (h *Handler) GetJSONHandler(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()

	var metric model.Metrics
	err := decoder.Decode(&metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.storage.GetMetric(&metric)

	if err != nil {
		if errors.Is(err, storage.ErrInvalidMetricType) {
			http.Error(res, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(res, err.Error(), http.StatusNotFound)
		}
		return
	}

	resp, err := json.Marshal(metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

func (h *Handler) UpdateJSONHandler(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()

	var metric model.Metrics
	err := decoder.Decode(&metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.storage.SaveMetric(metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.storage.GetMetric(&metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}
