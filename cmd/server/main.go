package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/handler"
	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/repository/memstorage"
	"github.com/derpartizanen/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	cfg := config.ConfigureServer()
	err := logger.Initialize(cfg.Loglevel)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		logger.Log.Error("can't connect to database")
	}
	defer db.Close()

	repository := memstorage.New()
	storageSettings := storage.Settings{
		StoragePath:   cfg.StoragePath,
		StoreInterval: cfg.StoreInterval,
		Restore:       cfg.Restore,
	}
	store := storage.New(repository, storageSettings, db)

	ctx := context.Background()
	if cfg.StoreInterval > 0 {
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
					if err := store.Backup(); err != nil {
						logger.Log.Error("Periodic backup failed", zap.Error(err))
					}
				}
			}
		}()
	}

	h := handler.NewHandler(store)
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(handler.GzipMiddleware)
	r.Get("/", h.GetAllHandler)
	r.Get("/value/{metricType}/{metricName}", h.GetHandler)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateHandler)
	r.Post("/value/", h.GetJSONHandler)
	r.Post("/update/", h.UpdateJSONHandler)
	r.Get("/ping", h.PingHandler)

	server := &http.Server{Addr: cfg.Host, Handler: r}
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
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Log.Fatal("Server shutdown failed", zap.Error(err))
		}
		logger.Log.Info("Server stopped gracefully")

		if err := store.Backup(); err != nil {
			logger.Log.Error("Metrics backup failed", zap.Error(err))
		}
		serverStopCtx()
	}()

	logger.Log.Info("Starting server on", zap.String("host", cfg.Host))
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal("Server quit unexpectedly", zap.Error(err))
	}

	<-serverCtx.Done()
}
