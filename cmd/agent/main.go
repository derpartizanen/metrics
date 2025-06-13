package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/derpartizanen/metrics/internal/agent"
	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/model"
	"github.com/derpartizanen/metrics/proto"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	err := logger.Initialize("INFO")
	if err != nil {
		log.Fatal(err)
	}
	logger.Log.Info("Agent", zap.String("version", buildVersion), zap.String("build_date", buildDate), zap.String("build_commit", buildCommit))

	cfg := config.ConfigureAgent()
	cfg.LogVars()

	httpClient := &http.Client{
		Timeout: time.Minute,
	}

	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatal("grpc client failed", zap.Error(err))
	}
	defer conn.Close()
	grpcClient := proto.NewMetricCollectorClient(conn)

	metricAgent := agent.New(httpClient, grpcClient, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	run(ctx, metricAgent)
}

func run(ctx context.Context, agent *agent.Agent) {
	pollTicker := time.NewTicker(time.Duration(agent.Config.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(agent.Config.ReportInterval) * time.Second)

	defer pollTicker.Stop()
	defer reportTicker.Stop()

	jobs := make(chan []model.Metrics, agent.Config.RateLimit)

	var wg sync.WaitGroup
	for i := 0; i <= agent.Config.RateLimit-1; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			agent.Worker(ctx, workerID, jobs)
		}(i + 1)
	}

	for {
		select {
		case <-pollTicker.C:
			agent.CollectMemStatsMetrics()
			agent.CollectPsutilMetrics()
		case <-reportTicker.C:
			agent.AddReportJob(ctx, jobs)
		case <-ctx.Done():
			logger.Log.Info("shutting down agent...")
			close(jobs)
			wg.Wait()
			return
		}
	}
}
