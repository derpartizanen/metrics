package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/derpartizanen/metrics/internal/compressor"
	"github.com/derpartizanen/metrics/internal/model"
)

func TestGzipMiddleware(t *testing.T) {
	metric := model.Metrics{ID: "test", MType: "counter"}
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var rmetric model.Metrics
		err := decoder.Decode(&rmetric)
		if err != nil {
			t.Error("can't decode response")
		}

		assert.Equal(t, metric, rmetric)
	})

	handlerToTest := GzipMiddleware(nextHandler)

	jsonStr, _ := json.Marshal(metric)
	gzipData, _ := compressor.Compress(jsonStr)
	req := httptest.NewRequest("POST", "http://test", bytes.NewBuffer(gzipData))
	req.Header.Set("Content-Encoding", "gzip")

	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
}
