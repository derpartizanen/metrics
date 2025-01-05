package main

import (
	"net/http"

	"github.com/derpartizanen/metrics/internal/handler"
	"github.com/derpartizanen/metrics/internal/repository/memstorage"
	"github.com/derpartizanen/metrics/internal/storage"
)

func main() {
	repository := memstorage.New()
	store := storage.New(repository)
	h := handler.NewHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
