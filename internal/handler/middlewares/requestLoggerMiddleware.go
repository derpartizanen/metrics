package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/derpartizanen/metrics/internal/logger"
)

// RequestLogger
// logs request info such as uri, method, size and so on
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 200,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)
		duration := time.Since(start)

		logger.Log.Info("request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Duration("duration", duration),
			zap.String("content-encoding", r.Header.Get("Content-Encoding")),
			zap.String("accept-encoding", r.Header.Get("Accept-Encoding")),
		)
	})
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write
// store response data size
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader
// store response status code
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
