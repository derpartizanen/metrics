package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/derpartizanen/metrics/internal/hash"
	"github.com/derpartizanen/metrics/internal/model"
	"github.com/derpartizanen/metrics/internal/repository/memstorage"
	"github.com/derpartizanen/metrics/internal/storage"
)

const HashHeader = "HashSHA256"

type Handler struct {
	storage *storage.Storage
	hashKey string
}

func NewHandler(storage *storage.Storage, hashKey string) *Handler {
	return &Handler{
		storage: storage,
		hashKey: hashKey,
	}
}

func (h *Handler) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")
	metricValue := req.PathValue("metricValue")

	err := h.storage.Save(metricType, metricName, metricValue)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (h *Handler) GetHandler(res http.ResponseWriter, req *http.Request) {
	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")

	var result string

	value, err := h.storage.Get(metricType, metricName)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidMetricType) {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, memstorage.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			http.Error(res, "metric not found", http.StatusNotFound)
			return
		}
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if metricType == model.MetricTypeCounter {
		result = fmt.Sprintf("%d", value)
	} else {
		result = fmt.Sprintf("%g", value)
	}

	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set(HashHeader, hash.Calc(h.hashKey, []byte(result)))
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, result)
}

func (h *Handler) GetAllHandler(res http.ResponseWriter, req *http.Request) {
	var result string
	metrics, _ := h.storage.GetAllMetrics()

	for _, metric := range metrics {
		if metric.MType == model.MetricTypeCounter {
			result += fmt.Sprintf("%s: %d\n", metric.ID, *metric.Delta)
		}
		if metric.MType == model.MetricTypeGauge {
			result += fmt.Sprintf("%s: %f\n", metric.ID, *metric.Value)
		}

	}
	res.Header().Set("Content-Type", "text/html")
	res.Header().Set(HashHeader, hash.Calc(h.hashKey, []byte(result)))
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
	res.Header().Set(HashHeader, hash.Calc(h.hashKey, resp))
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

func (h *Handler) BatchUpdateJSONHandler(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()

	var metrics []model.Metrics
	err := decoder.Decode(&metrics)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.storage.SetAllMetrics(metrics)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (h *Handler) PingHandler(res http.ResponseWriter, req *http.Request) {
	err := h.storage.Ping()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.WriteHeader(http.StatusOK)
}
