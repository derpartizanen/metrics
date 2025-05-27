// Package agent contains methods for collecting and sending metrics to server
package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	"github.com/derpartizanen/metrics/internal/compressor"
	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/crypto"
	"github.com/derpartizanen/metrics/internal/hash"
	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/model"
)

type Agent struct {
	Config  *config.AgentConfig
	PubKey  []byte
	Client  *http.Client
	Metrics []model.Metrics
	mu      sync.Mutex
}

var (
	counter      int64
	ErrDoRequest = errors.New("execution request error")
)

func New(client *http.Client, config *config.AgentConfig) *Agent {
	var pubKey []byte
	var err error
	if config.CryptoKey != "" {
		pubKey, err = crypto.ReadKeyFile(config.CryptoKey)
		if err != nil {
			logger.Log.Fatal("read public key", zap.String("error", err.Error()))
		}
	}
	return &Agent{Client: client, Metrics: []model.Metrics{}, Config: config, PubKey: pubKey}
}

// CollectPsutilMetrics
// Collect mem.VirtualMemory's Total, Free, UsedPercent values
func (agent *Agent) CollectPsutilMetrics() {
	vm, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error("virtual memory error", zap.Error(err))
		return
	}

	totalMemory := float64(vm.Total)
	freeMemory := float64(vm.Free)
	CPUutilization := vm.UsedPercent

	agent.SetGaugeMetric("TotalMemory", &totalMemory)
	agent.SetGaugeMetric("FreeMemory", &freeMemory)
	agent.SetGaugeMetric("CPUutilization1", &CPUutilization)
}

// CollectMemStatsMetrics
// Collect memory stats from runtime
func (agent *Agent) CollectMemStatsMetrics() {
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)
	mValue := reflect.ValueOf(memStats)
	mType := mValue.Type()

	for _, metricName := range model.GaugeMetrics {
		field, ok := mType.FieldByName(metricName)
		if !ok {
			continue
		}

		var value float64

		switch mValue.FieldByName(metricName).Interface().(type) {
		case uint64:
			value = float64(mValue.FieldByName(metricName).Interface().(uint64))
		case uint32:
			value = float64(mValue.FieldByName(metricName).Interface().(uint32))
		case float64:
			value = mValue.FieldByName(metricName).Interface().(float64)
		}

		agent.SetGaugeMetric(field.Name, &value)
	}

	counter += 1
	agent.SetCounterMetric("PollCount", &counter)

	random := rand.Float64()
	agent.SetGaugeMetric("RandomValue", &random)
}

// SetGaugeMetric
// Update gauge metric and append to metrics slice
func (agent *Agent) SetGaugeMetric(metricName string, metricValue *float64) {
	agent.mu.Lock()
	defer agent.mu.Unlock()

	agent.Metrics = append(agent.Metrics, model.Metrics{ID: metricName, MType: model.MetricTypeGauge, Value: metricValue})
}

// SetCounterMetric
// Update counter metric and append to metrics slice
func (agent *Agent) SetCounterMetric(metricName string, metricDelta *int64) {
	agent.mu.Lock()
	defer agent.mu.Unlock()

	agent.Metrics = append(agent.Metrics, model.Metrics{ID: metricName, MType: model.MetricTypeCounter, Delta: metricDelta})
}

// AddReportJob
// Sends collected metrics to jobs channel
func (agent *Agent) AddReportJob(jobs chan<- []model.Metrics) {
	jobs <- agent.Metrics
}

// Worker
// reads metrics from jobs channel and sends them to server with retries
func (agent *Agent) Worker(ctx context.Context, id int, jobs <-chan []model.Metrics) {
	for metrics := range jobs {
		logger.Log.Debug(fmt.Sprintf("start job on worker %d", id))

		err := agent.reportMetricsWithRetry(ctx, metrics)
		if err != nil {
			logger.Log.Error("send request error", zap.Error(err))
		}

		logger.Log.Debug(fmt.Sprintf("finish job on worker %d", id))
	}
}

func (agent *Agent) reportMetricsWithRetry(ctx context.Context, metrics []model.Metrics) error {
	var err error
	for i := 1; i <= agent.Config.ReportRetryCount; i++ {
		err = agent.reportMetrics(ctx, metrics)
		if i == agent.Config.ReportRetryCount {
			break
		}
		if err != nil && errors.Is(err, ErrDoRequest) {
			logger.Log.Info(fmt.Sprintf("retry %d to report metrics", i))
			time.Sleep(time.Duration(i+i-1) * time.Second)
			continue
		}
		break
	}

	if err != nil {
		logger.Log.Error("failed to report metrics", zap.Error(err))
	}

	return err
}

func (agent *Agent) reportMetrics(ctx context.Context, metrics []model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	reportURL := fmt.Sprintf("http://%s/updates/", agent.Config.Address)

	jsonStr, err := json.Marshal(metrics)
	if err != nil {
		return errors.New("can't marshal data")
	}

	gzipData, err := compressor.Compress(jsonStr)
	if err != nil {
		return errors.New("can't compress data")
	}
	b := gzipData
	if agent.Config.CryptoKey != "" {
		encodedBytes, err := crypto.Encrypt(gzipData, agent.PubKey)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}
		b = encodedBytes
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reportURL, bytes.NewBuffer(b))
	if err != nil {
		return errors.New("new request error")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")

	if agent.Config.HashKey != "" {
		requestHash := hash.Calc(agent.Config.HashKey, jsonStr)
		req.Header.Add("HashSHA256", requestHash)
	}

	res, err := agent.Client.Do(req)
	if err != nil {
		return ErrDoRequest
	}

	logger.Log.Debug(fmt.Sprintf("send batch request with %d metrics", len(metrics)))
	res.Body.Close()

	return nil
}
