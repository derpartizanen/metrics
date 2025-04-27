package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/derpartizanen/metrics/internal/handler"
	"github.com/derpartizanen/metrics/internal/hash"
)

// VerifyHash
// check that hash in the header is equal with hash computing for body data
func (hm *HashMiddleware) VerifyHash(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(handler.HashHeader) != "" {
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(payload))

			hashStr := hash.Calc(hm.HashKey, payload)
			if hashStr != r.Header.Get(handler.HashHeader) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}

type HashMiddleware struct {
	HashKey string
}

func NewHashMiddleware(hash string) *HashMiddleware {
	return &HashMiddleware{
		HashKey: hash,
	}
}
