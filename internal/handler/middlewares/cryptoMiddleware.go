package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/derpartizanen/metrics/internal/crypto"
	"github.com/derpartizanen/metrics/internal/logger"
)

func (cm *CryptoMiddleware) Decrypt() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			if len(cm.PrivateKey) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error("Decrypt", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			body, err = crypto.Decrypt(body, cm.PrivateKey)
			if err != nil {
				logger.Log.Error("Decrypt", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(body))

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

type CryptoMiddleware struct {
	PrivateKey []byte
}

func NewCryptoMiddleware(privateKeyPath string) *CryptoMiddleware {
	var privateKey []byte
	var err error
	if privateKeyPath != "" {
		privateKey, err = crypto.ReadKeyFile(privateKeyPath)
		if err != nil {
			logger.Log.Fatal("readKeyFile", zap.Error(err))
		}
	}

	return &CryptoMiddleware{
		PrivateKey: privateKey,
	}
}
