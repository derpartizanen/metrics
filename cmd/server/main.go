package main

import (
	"log"
	"net/http"

	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/handler"
	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/repository/memstorage"
	"github.com/derpartizanen/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	cfg := config.ConfigureServer()
	err := logger.Initialize("INFO")
	if err != nil {
		log.Fatal(err)
	}

	repository := memstorage.New()
	store := storage.New(repository)
	h := handler.NewHandler(store)

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Get("/", h.GetAllHandler)
	r.Get("/value/{metricType}/{metricName}", h.GetHandler)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler)

	logger.Log.Info("Starting server on", zap.String("host", cfg.Host))
	err = http.ListenAndServe(cfg.Host, r)
	if err != nil {
		panic(err)
	}
}
