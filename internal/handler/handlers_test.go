package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/model"
	"github.com/derpartizanen/metrics/internal/storage"
)

func TestHandler_UpdateHandler(t *testing.T) {
	var baseURL = "http://localhost:8080"
	cfg := config.ConfigureServer()
	store := storage.New(context.Background(), *cfg)
	h := NewHandler(store, cfg.Key)

	tests := []struct {
		name         string
		method       string
		endpoint     string
		expectedCode int
	}{
		{name: "gauge request", method: http.MethodPost, endpoint: "/update/gauge/Alloc/123", expectedCode: 200},
		{name: "counter request", method: http.MethodPost, endpoint: "/update/counter/PollCounter/2", expectedCode: 200},
		{name: "bad gauge request", method: http.MethodPost, endpoint: "/update/gauge/Alloc/test", expectedCode: 400},
		{name: "bad counter request", method: http.MethodPost, endpoint: "/update/counter/PollCounter/3.14", expectedCode: 400},
		{name: "bad metric type", method: http.MethodPost, endpoint: "/update/bad/Alloc/123", expectedCode: 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, fmt.Sprintf("%s%s", baseURL, tt.endpoint), nil)
			req.SetPathValue("metricType", strings.Split(tt.endpoint, "/")[2])
			req.SetPathValue("metricName", strings.Split(tt.endpoint, "/")[3])
			req.SetPathValue("metricValue", strings.Split(tt.endpoint, "/")[4])
			res := httptest.NewRecorder()

			h.UpdateHandler(res, req)
			assert.Equal(t, tt.expectedCode, res.Code)
		})
	}
}

func TestHandler_UpdateJSONHandler(t *testing.T) {
	var baseURL = "http://localhost:8080"
	cfg := config.ServerConfig{}
	store := storage.New(context.Background(), cfg)
	h := NewHandler(store, cfg.Key)

	tests := []struct {
		name         string
		method       string
		endpoint     string
		expectedCode int
		payload      string
	}{
		{
			name:         "gauge metric",
			method:       http.MethodPost,
			endpoint:     "/update/",
			expectedCode: 200,
			payload:      `{"id":"Alloc","type":"gauge","value": 123}`,
		},
		{
			name:         "counter metric",
			method:       http.MethodPost,
			endpoint:     "/update/",
			expectedCode: 200,
			payload:      `{"id":"PollCounter","type":"counter","delta": 10}`,
		},
		{
			name:         "bad gauge",
			method:       http.MethodPost,
			endpoint:     "/update/",
			expectedCode: 400,
			payload:      `{"id":"Alloc","type":"gauge","value": "abc"}`,
		},
		{
			name:         "bad counter",
			method:       http.MethodPost,
			endpoint:     "/update/",
			expectedCode: 400,
			payload:      `{"id":"PollCounter","type":"counter","delta": 10.33}`,
		},
		{
			name:         "bad metric type",
			method:       http.MethodPost,
			endpoint:     "/update/",
			expectedCode: 400,
			payload:      `{"id":"Alloc","type":"bad","value": 123}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, fmt.Sprintf("%s%s", baseURL, tt.endpoint), bytes.NewBuffer([]byte(tt.payload)))
			res := httptest.NewRecorder()

			h.UpdateJSONHandler(res, req)
			assert.Equal(t, tt.expectedCode, res.Code)
		})
	}
}

func TestHandler_BatchUpdateJSONHandler(t *testing.T) {
	var baseURL = "http://localhost:8080"
	cfg := config.ServerConfig{}
	store := storage.New(context.Background(), cfg)
	h := NewHandler(store, cfg.Key)

	tests := []struct {
		name         string
		method       string
		endpoint     string
		expectedCode int
		payload      string
	}{
		{
			name:         "single metric",
			method:       http.MethodPost,
			endpoint:     "/updates/",
			expectedCode: 200,
			payload:      `[{"id":"Alloc","type":"gauge","value": 123}]`,
		},
		{
			name:         "multiple metrics",
			method:       http.MethodPost,
			endpoint:     "/updates/",
			expectedCode: 200,
			payload:      `[{"id":"PollCounter","type":"counter","delta": 10}, {"id":"Alloc","type":"gauge","value": 123}]`,
		},
		{
			name:         "invalid json",
			method:       http.MethodPost,
			endpoint:     "/updates/",
			expectedCode: 400,
			payload:      `[{id:"Alloc","type":"gauge","value": "abc"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, fmt.Sprintf("%s%s", baseURL, tt.endpoint), bytes.NewBuffer([]byte(tt.payload)))
			res := httptest.NewRecorder()

			h.BatchUpdateJSONHandler(res, req)
			assert.Equal(t, tt.expectedCode, res.Code)
		})
	}
}

func TestHandler_GetHandler(t *testing.T) {
	var baseURL = "http://localhost:8080"
	cfg := config.ServerConfig{}
	store := storage.New(context.Background(), cfg)
	h := NewHandler(store, cfg.Key)

	store.Save(model.MetricTypeGauge, "Alloc", "123")
	store.Save(model.MetricTypeCounter, "PollCounter", "10")

	tests := []struct {
		name           string
		method         string
		endpoint       string
		expectedCode   int
		expectedResult string
	}{
		{
			name:           "gauge exists",
			method:         http.MethodGet,
			endpoint:       "/value/gauge/Alloc",
			expectedCode:   200,
			expectedResult: "123",
		},
		{
			name:         "gauge missed",
			method:       http.MethodGet,
			endpoint:     "/value/gauge/Mallocs",
			expectedCode: 404,
		},
		{
			name:           "counter exists",
			method:         http.MethodGet,
			endpoint:       "/value/counter/PollCounter",
			expectedCode:   200,
			expectedResult: "10",
		},
		{
			name:         "counter missed",
			method:       http.MethodGet,
			endpoint:     "/value/counter/Counter",
			expectedCode: 404,
		},
		{
			name:         "invalid type",
			method:       http.MethodGet,
			endpoint:     "/value/wrongtype/Counter",
			expectedCode: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, fmt.Sprintf("%s%s", baseURL, tt.endpoint), nil)
			req.SetPathValue("metricType", strings.Split(tt.endpoint, "/")[2])
			req.SetPathValue("metricName", strings.Split(tt.endpoint, "/")[3])
			res := httptest.NewRecorder()

			h.GetHandler(res, req)
			assert.Equal(t, tt.expectedCode, res.Code)
			if res.Code == 200 {
				readBuf, _ := io.ReadAll(res.Body)
				assert.Equal(t, tt.expectedResult, string(readBuf))
			}
		})
	}
}

func TestHandler_GetJSONHandler(t *testing.T) {
	var baseURL = "http://localhost:8080"
	cfg := config.ServerConfig{}
	store := storage.New(context.Background(), cfg)
	h := NewHandler(store, cfg.Key)

	store.Save(model.MetricTypeGauge, "Alloc", "123")
	store.Save(model.MetricTypeCounter, "PollCounter", "10")

	tests := []struct {
		name           string
		method         string
		endpoint       string
		payload        string
		expectedCode   int
		expectedResult string
	}{
		{
			name:           "gauge exists",
			method:         http.MethodPost,
			endpoint:       "/value/",
			payload:        `{"id":"Alloc", "type":"gauge"}`,
			expectedCode:   200,
			expectedResult: `{"id":"Alloc","type":"gauge","value": 123}`,
		},
		{
			name:           "counter exists",
			method:         http.MethodPost,
			endpoint:       "/value/",
			payload:        `{"id":"PollCounter", "type":"counter"}`,
			expectedCode:   200,
			expectedResult: `{"id":"PollCounter","type":"counter","delta": 10}`,
		},
		{
			name:         "gauge not found",
			method:       http.MethodPost,
			endpoint:     "/value/",
			payload:      `{"id":"MAlloc", "type":"gauge"}`,
			expectedCode: 404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, fmt.Sprintf("%s%s", baseURL, tt.endpoint), bytes.NewBuffer([]byte(tt.payload)))
			res := httptest.NewRecorder()

			h.GetJSONHandler(res, req)
			assert.Equal(t, tt.expectedCode, res.Code)
			if res.Code == 200 {
				readBuf, _ := io.ReadAll(res.Body)
				assert.JSONEq(t, tt.expectedResult, string(readBuf))
			}
		})
	}
}
