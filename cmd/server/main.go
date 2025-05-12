package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/handler"
	middlewares "github.com/derpartizanen/metrics/internal/handler/middlewares"
	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/server"
	"github.com/derpartizanen/metrics/internal/storage"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	cfg := config.ConfigureServer()
	err := logger.Initialize(cfg.Loglevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.Log.Info("Server", zap.String("version", buildVersion), zap.String("build_date", buildDate), zap.String("build_commit", buildCommit))

	ctx := context.Background()
	store := storage.New(ctx, cfg)

	if cfg.DatabaseDSN == "" && cfg.StoreInterval > 0 {
		logger.Log.Debug(fmt.Sprintf("Activate periodic backups with interval %d seconds", cfg.StoreInterval))
		go func() {
			ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					logger.Log.Debug("Running periodic backup")
					if backupErr := store.Backup(); backupErr != nil {
						logger.Log.Error("Periodic backup failed", zap.Error(err))
					}
				}
			}
		}()
	}

	h := handler.NewHandler(store, cfg.Key)
	r := chi.NewRouter()
	r.Use(middlewares.RequestLogger)
	r.Use(middlewares.GzipMiddleware)
	hm := middlewares.NewHashMiddleware(cfg.Key)
	r.Use(hm.VerifyHash)
	r.Mount("/debug", middleware.Profiler())
	r.Get("/", h.GetAllHandler)
	r.Get("/value/{metricType}/{metricName}", h.GetHandler)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler)
	r.Post("/value/", h.GetJSONHandler)
	r.Post("/update/", h.UpdateJSONHandler)
	r.Post("/updates/", h.BatchUpdateJSONHandler)
	r.Get("/ping", h.PingHandler)

	srv := server.New(cfg.Host, r)
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, cancel := context.WithTimeout(serverCtx, 10*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Log.Fatal("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		logger.Log.Info("Shutting down server...")
		shutdownErr := srv.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			logger.Log.Fatal("Server shutdown failed", zap.Error(err))
		}
		logger.Log.Info("Server stopped gracefully")

		if cfg.DatabaseDSN == "" {
			if backupErr := store.Backup(); backupErr != nil {
				logger.Log.Error("Metrics backup failed", zap.Error(err))
			}
		}
		serverStopCtx()
	}()

	logger.Log.Info("Starting server on", zap.String("host", cfg.Host))
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal("Server quit unexpectedly", zap.Error(err))
	}

	<-serverCtx.Done()
}
