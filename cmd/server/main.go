package main

import (
	"net/http"

	"github.com/derpartizanen/metrics/internal/handler"
	"github.com/derpartizanen/metrics/internal/repository/memstorage"
	"github.com/derpartizanen/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	repository := memstorage.New()
	store := storage.New(repository)
	h := handler.NewHandler(store)

	r := chi.NewRouter()
	r.Get("/", h.GetAllHandler)
	r.Get("/value/{metricType}/{metricName}", h.GetHandler)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
