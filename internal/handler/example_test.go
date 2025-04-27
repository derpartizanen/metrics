package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/model"
	"github.com/derpartizanen/metrics/internal/storage"
)

func ExampleHandler_GetHandler() {
	cfg := config.ServerConfig{}
	service := storage.New(context.Background(), cfg)
	h := NewHandler(service, cfg.Key)

	// prepare storage data
	value := 1.25
	metric := model.Metrics{
		ID:    "Alloc",
		MType: model.MetricTypeGauge,
		Value: &value,
	}
	service.SaveMetric(metric)

	mux := chi.NewRouter()
	mux.Get("/value/{metricType}/{metricName}", h.GetHandler)
	s := httptest.NewServer(mux)
	defer s.Close()

	endpoint := "/value/gauge/Alloc"
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, s.URL+endpoint, http.NoBody,
	)

	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	res, err := s.Client().Do(req)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	io.Copy(os.Stdout, res.Body)

	// Output:
	// 200
	// 1.25
}

func ExampleHandler_GetJSONHandler() {
	cfg := config.ServerConfig{}
	service := storage.New(context.Background(), cfg)
	h := NewHandler(service, cfg.Key)

	// prepare storage data
	value := 1.25
	metric := model.Metrics{
		ID:    "Alloc",
		MType: model.MetricTypeGauge,
		Value: &value,
	}
	service.SaveMetric(metric)

	mux := chi.NewRouter()
	mux.Post("/value/", h.GetJSONHandler)
	s := httptest.NewServer(mux)
	defer s.Close()

	var reqData bytes.Buffer
	json.NewEncoder(&reqData).Encode(model.Metrics{ID: "Alloc", MType: model.MetricTypeGauge})
	endpoint := "/value/"
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, s.URL+endpoint, &reqData,
	)

	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	res, err := s.Client().Do(req)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	io.Copy(os.Stdout, res.Body)

	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":1.25}
}

func ExampleHandler_GetAllHandler() {
	cfg := config.ServerConfig{}
	service := storage.New(context.Background(), cfg)
	h := NewHandler(service, cfg.Key)

	// prepare storage data
	service.Save(model.MetricTypeGauge, "Alloc", "1.250000")
	service.Save(model.MetricTypeCounter, "Count", "10")

	mux := chi.NewRouter()
	mux.Get("/", h.GetAllHandler)
	s := httptest.NewServer(mux)
	defer s.Close()

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, s.URL, http.NoBody,
	)

	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	res, err := s.Client().Do(req)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	io.Copy(os.Stdout, res.Body)

	// Output:
	// 200
	// Alloc: 1.250000
	// Count: 10
}

func ExampleHandler_UpdateHandler() {
	cfg := config.ServerConfig{}
	service := storage.New(context.Background(), cfg)
	h := NewHandler(service, cfg.Key)

	mux := chi.NewRouter()
	mux.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler)
	s := httptest.NewServer(mux)
	defer s.Close()

	endpoint := "/update/gauge/Alloc/1.23"
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, s.URL+endpoint, http.NoBody,
	)

	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	res, err := s.Client().Do(req)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	io.Copy(os.Stdout, res.Body)

	// Output:
	// 200
}

func ExampleHandler_UpdateJSONHandler() {
	cfg := config.ServerConfig{}
	service := storage.New(context.Background(), cfg)
	h := NewHandler(service, cfg.Key)

	mux := chi.NewRouter()
	mux.Post("/update/", h.UpdateJSONHandler)
	s := httptest.NewServer(mux)
	defer s.Close()

	value := 1.25
	metric := model.Metrics{
		ID:    "Alloc",
		MType: model.MetricTypeGauge,
		Value: &value,
	}

	var reqData bytes.Buffer
	if err := json.NewEncoder(&reqData).Encode(metric); err != nil {
		logger.Log.Fatal(err.Error())
	}

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, s.URL+"/update/", &reqData,
	)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := s.Client().Do(req)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	io.Copy(os.Stdout, res.Body)

	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":1.25}
}

func ExampleHandler_BatchUpdateJSONHandler() {
	cfg := config.ServerConfig{}
	service := storage.New(context.Background(), cfg)
	h := NewHandler(service, cfg.Key)

	mux := chi.NewRouter()
	mux.Post("/updates/", h.BatchUpdateJSONHandler)
	s := httptest.NewServer(mux)
	defer s.Close()

	value := 1.25
	gMetric := model.Metrics{
		ID:    "Alloc",
		MType: model.MetricTypeGauge,
		Value: &value,
	}

	delta := int64(100)
	cMetric := model.Metrics{
		ID:    "Count",
		MType: model.MetricTypeCounter,
		Delta: &delta,
	}

	var metrics []model.Metrics
	metrics = append(metrics, gMetric, cMetric)

	var reqData bytes.Buffer
	if err := json.NewEncoder(&reqData).Encode(metrics); err != nil {
		logger.Log.Fatal(err.Error())
	}

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, s.URL+"/updates/", &reqData,
	)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := s.Client().Do(req)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	io.Copy(os.Stdout, res.Body)

	// Output:
	// 200
}
