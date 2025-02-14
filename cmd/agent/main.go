package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/model"
	"go.uber.org/zap"
)

var (
	ErrDoRequest = errors.New("execution request error")
)

var cfg *config.AgentConfig
var counter int64

func main() {
	var metrics []model.Metrics

	err := logger.Initialize("INFO")
	if err != nil {
		log.Fatal(err)
	}
	logger.Log.Info("Starting agent")

	cfg = config.ConfigureAgent()
	cfg.LogVars()

	go func() {
		for {
			metrics = updateMetrics()
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		}
	}()

	for {
		time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
		reportMetricsWithRetry(metrics, 3)
	}
}

func updateMetrics() []model.Metrics {
	var metrics []model.Metrics
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
		default:
			return nil
		}

		metrics = append(metrics, model.Metrics{ID: field.Name, MType: "gauge", Value: &value})
	}

	counter += 1
	metrics = append(metrics, model.Metrics{ID: "PollCount", MType: "counter", Delta: &counter})
	random := rand.Float64()
	metrics = append(metrics, model.Metrics{ID: "RandomValue", MType: "gauge", Value: &random})

	return metrics
}

func reportMetricsWithRetry(metrics []model.Metrics, retryCount int) {
	var err error
	for i := 1; i <= retryCount; i++ {
		err = reportMetrics(metrics)
		if i == retryCount {
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
}

func reportMetrics(metrics []model.Metrics) error {
	if len(metrics) == 0 {
		return errors.New("empty metrics")
	}

	reportURL := fmt.Sprintf("http://%s/updates/", cfg.ReportEndpoint)
	client := &http.Client{}

	jsonStr, err := json.Marshal(metrics)
	if err != nil {
		return errors.New("can't marshal data")
	}

	gzipData, err := compressData(jsonStr)
	if err != nil {
		return errors.New("can't compress data")
	}

	req, err := http.NewRequest(http.MethodPost, reportURL, bytes.NewBuffer(gzipData))
	if err != nil {
		return errors.New("new request error")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")
	res, err := client.Do(req)
	if err != nil {
		return ErrDoRequest
	}

	logger.Log.Debug(fmt.Sprintf("send batch request with %d metrics", len(metrics)))
	res.Body.Close()

	return nil
}

func compressData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}

	_, err = w.Write(data)
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
