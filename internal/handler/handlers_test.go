package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/derpartizanen/metrics/internal/repository/memstorage"
	"github.com/derpartizanen/metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestHandler_UpdateHandler(t *testing.T) {
	var baseUrl = "http://localhost:8080"
	repository := memstorage.New()
	store := storage.New(repository)
	h := NewHandler(store)

	tests := []struct {
		name         string
		method       string
		endpoint     string
		expectedCode int
	}{
		{name: "gauge requqest", method: http.MethodPost, endpoint: "/update/gauge/Alloc/123", expectedCode: 200},
		{name: "counter reqeust", method: http.MethodPost, endpoint: "/update/counter/PollCounter/2", expectedCode: 200},
		{name: "bad gauge request", method: http.MethodPost, endpoint: "/update/gauge/Alloc/test", expectedCode: 400},
		{name: "bad counter request", method: http.MethodPost, endpoint: "/update/counter/PollCounter/3.14", expectedCode: 400},
		{name: "bad metric type", method: http.MethodPost, endpoint: "/update/bad/Alloc/123", expectedCode: 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, fmt.Sprintf("%s%s", baseUrl, tt.endpoint), nil)
			req.SetPathValue("metricType", strings.Split(tt.endpoint, "/")[2])
			req.SetPathValue("metricName", strings.Split(tt.endpoint, "/")[3])
			req.SetPathValue("metricValue", strings.Split(tt.endpoint, "/")[4])
			res := httptest.NewRecorder()

			h.UpdateHandler(res, req)
			assert.Equal(t, tt.expectedCode, res.Code)
		})
	}
}
