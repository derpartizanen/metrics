package router

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/handler"
	"github.com/derpartizanen/metrics/internal/handler/middlewares"
)

func NewRouter(cfg *config.ServerConfig, h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()

	if len(cfg.CryptoKey) > 0 {
		cm := middlewares.NewCryptoMiddleware(cfg.CryptoKey)
		r.Use(cm.Decrypt())
	}

	if cfg.TrustedSubnet != "" {
		tm := middlewares.NewTrustedSubnetMiddleware(cfg.TrustedSubnet)
		r.Use(tm.VerifySubnet)
	}

	r.Use(middlewares.RequestLogger)
	r.Use(middlewares.GzipMiddleware)
	hm := middlewares.NewHashMiddleware(cfg.Key)
	r.Use(hm.VerifyHash)

	r.Mount("/debug", chimiddleware.Profiler())
	r.Get("/", h.GetAllHandler)
	r.Get("/value/{metricType}/{metricName}", h.GetHandler)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler)
	r.Post("/value/", h.GetJSONHandler)
	r.Post("/update/", h.UpdateJSONHandler)
	r.Post("/updates/", h.BatchUpdateJSONHandler)
	r.Get("/ping", h.PingHandler)

	return r
}
