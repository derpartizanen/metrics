package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/derpartizanen/metrics/internal/agent"
	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/model"
)

func main() {
	err := logger.Initialize("INFO")
	if err != nil {
		log.Fatal(err)
	}
	logger.Log.Info("Starting agent")

	cfg := config.ConfigureAgent()
	cfg.LogVars()

	client := &http.Client{
		Timeout: time.Minute,
	}

	metricAgent := agent.New(client, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	run(ctx, metricAgent)
}

func run(ctx context.Context, agent *agent.Agent) {

	pollTicker := time.NewTicker(time.Duration(agent.Config.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(agent.Config.ReportInterval) * time.Second)

	defer pollTicker.Stop()
	defer reportTicker.Stop()

	jobs := make(chan []model.Metrics, agent.Config.RateLimit)
	defer close(jobs)

	for i := 0; i < agent.Config.RateLimit; i++ {
		go agent.Worker(ctx, i+1, jobs)
	}

	for {
		select {
		case <-pollTicker.C:
			go agent.CollectMemStatsMetrics()
			go agent.CollectPsutilMetrics()
		case <-reportTicker.C:
			go agent.AddReportJob(jobs)
		case <-ctx.Done():
			logger.Log.Info("shutting down agent...")
			return
		}
	}
}
