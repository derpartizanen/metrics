package handler

import (
	"net/http"

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
