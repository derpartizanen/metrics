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

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	grpcserver "github.com/derpartizanen/metrics/grpc/server"
	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/handler"
	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/network"
	"github.com/derpartizanen/metrics/internal/router"
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
	cfg.LogVars()
	err := logger.Initialize(cfg.Loglevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.Log.Info("Server", zap.String("version", buildVersion), zap.String("build_date", buildDate), zap.String("build_commit", buildCommit))

	ctx, cancel := context.WithCancel(context.Background())
	service := storage.New(ctx, *cfg)

	if cfg.DatabaseDSN == "" && cfg.StoreInterval > 0 {
		go runPeriodicBackups(ctx, cfg, service)
	}

	h := handler.NewHandler(service, cfg.Key)
	r := router.NewRouter(cfg, h)

	httpServer := server.New(cfg.Host, r)
	go func() {
		defer cancel()
		logger.Log.Info("Starting server on", zap.String("host", cfg.Host))
		err = httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server quit unexpectedly", zap.Error(err))
		}
	}()

	grpcServer := grpcserver.NewServer(grpcserver.Config{
		ServerAddr:    cfg.GrpcAddress,
		Service:       service,
		TrustedSubnet: network.GetSubnetFromString(cfg.TrustedSubnet),
	})

	go func() {
		defer cancel()
		if err := grpcServer.Run(ctx); err != nil {
			logger.Log.Fatal("can't start server", zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case <-sig:
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			logger.Log.Info("Shutting down http server...")
			shutdownErr := httpServer.Shutdown(ctx)
			if shutdownErr != nil {
				logger.Log.Fatal("http server shutdown failed", zap.Error(shutdownErr))
			}
			logger.Log.Info("http server stopped gracefully")

			if err := grpcServer.Shutdown(ctx); err != nil {
				logger.Log.Fatal("grpc server shutdown failed", zap.Error(err))
			}
			logger.Log.Info("grpc server stopped gracefully")

			if cfg.DatabaseDSN == "" {
				if backupErr := service.Backup(); backupErr != nil {
					logger.Log.Error("Metrics backup failed", zap.Error(err))
				}
			}
			return
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				logger.Log.Info("Graceful shutdown timed out.. forcing exit.")
			}
			return
		}
	}
}

func runPeriodicBackups(ctx context.Context, cfg *config.ServerConfig, service *storage.Storage) {
	logger.Log.Debug(fmt.Sprintf("Activate periodic backups with interval %d seconds", cfg.StoreInterval))

	ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logger.Log.Debug("Running periodic backup")
			if backupErr := service.Backup(); backupErr != nil {
				logger.Log.Error("Periodic backup failed", zap.Error(backupErr))
			}
		}
	}
}
