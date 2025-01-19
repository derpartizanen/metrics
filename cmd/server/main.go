package main

import (
	"log"
	"net/http"

	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/handler"
	"github.com/derpartizanen/metrics/internal/repository/memstorage"
	"github.com/derpartizanen/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.ConfigureServer()
	repository := memstorage.New()
	store := storage.New(repository)
	h := handler.NewHandler(store)

	r := chi.NewRouter()
	r.Get("/", h.GetAllHandler)
	r.Get("/value/{metricType}/{metricName}", h.GetHandler)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler)

	log.Println("Starting server on", cfg.Host)
	err := http.ListenAndServe(cfg.Host, r)
	if err != nil {
		panic(err)
	}
}
