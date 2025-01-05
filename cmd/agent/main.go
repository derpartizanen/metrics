package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/derpartizanen/metrics/internal/model"
)

var (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	reportUrl      = "http://localhost:8080/update"
	counter        int64
)

func main() {
	var metrics []model.Metric

	go func() {
		for {
			metrics = updateMetrics()
			time.Sleep(pollInterval)
		}
	}()

	for {
		time.Sleep(reportInterval)
		err := reportMetrics(metrics)
		if err != nil {
			log.Fatal("report metric error: ", err)
		}
	}
}

func updateMetrics() []model.Metric {
	var metrics []model.Metric

	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	metrics = append(metrics, model.Metric{Name: "Alloc", Type: "gauge", Value: float64(m.Alloc)})
	metrics = append(metrics, model.Metric{Name: "TotalAlloc", Type: "gauge", Value: float64(m.TotalAlloc)})
	metrics = append(metrics, model.Metric{Name: "Sys", Type: "gauge", Value: float64(m.Sys)})
	metrics = append(metrics, model.Metric{Name: "Lookups", Type: "gauge", Value: float64(m.Lookups)})
	metrics = append(metrics, model.Metric{Name: "Mallocs", Type: "gauge", Value: float64(m.Mallocs)})
	metrics = append(metrics, model.Metric{Name: "Frees", Type: "gauge", Value: float64(m.Frees)})
	metrics = append(metrics, model.Metric{Name: "HeapAlloc", Type: "gauge", Value: float64(m.HeapAlloc)})
	metrics = append(metrics, model.Metric{Name: "HeapSys", Type: "gauge", Value: float64(m.HeapSys)})
	metrics = append(metrics, model.Metric{Name: "HeapIdle", Type: "gauge", Value: float64(m.HeapIdle)})
	metrics = append(metrics, model.Metric{Name: "HeapInuse", Type: "gauge", Value: float64(m.HeapInuse)})
	metrics = append(metrics, model.Metric{Name: "HeapReleased", Type: "gauge", Value: float64(m.HeapReleased)})
	metrics = append(metrics, model.Metric{Name: "HeapObjects", Type: "gauge", Value: float64(m.HeapObjects)})
	metrics = append(metrics, model.Metric{Name: "StackInuse", Type: "gauge", Value: float64(m.StackInuse)})
	metrics = append(metrics, model.Metric{Name: "StackSys", Type: "gauge", Value: float64(m.StackSys)})
	metrics = append(metrics, model.Metric{Name: "MSpanInuse", Type: "gauge", Value: float64(m.MSpanInuse)})
	metrics = append(metrics, model.Metric{Name: "MSpanSys", Type: "gauge", Value: float64(m.MSpanSys)})
	metrics = append(metrics, model.Metric{Name: "MCacheInuse", Type: "gauge", Value: float64(m.MCacheInuse)})
	metrics = append(metrics, model.Metric{Name: "MCacheSys", Type: "gauge", Value: float64(m.MCacheSys)})
	metrics = append(metrics, model.Metric{Name: "BuckHashSys", Type: "gauge", Value: float64(m.BuckHashSys)})
	metrics = append(metrics, model.Metric{Name: "GCSys", Type: "gauge", Value: float64(m.GCSys)})
	metrics = append(metrics, model.Metric{Name: "OtherSys", Type: "gauge", Value: float64(m.OtherSys)})
	metrics = append(metrics, model.Metric{Name: "NextGC", Type: "gauge", Value: float64(m.NextGC)})
	metrics = append(metrics, model.Metric{Name: "LastGC", Type: "gauge", Value: float64(m.LastGC)})
	metrics = append(metrics, model.Metric{Name: "PauseTotalNs", Type: "gauge", Value: float64(m.PauseTotalNs)})
	metrics = append(metrics, model.Metric{Name: "GCCPUFraction", Type: "gauge", Value: float64(m.GCCPUFraction)})
	metrics = append(metrics, model.Metric{Name: "NumForcedGC", Type: "gauge", Value: float64(m.NumForcedGC)})
	metrics = append(metrics, model.Metric{Name: "NumGC", Type: "gauge", Value: float64(m.NumGC)})

	counter += 1
	metrics = append(metrics, model.Metric{Name: "RandomValue", Type: "gauge", Value: rand.Float64()})
	metrics = append(metrics, model.Metric{Name: "PollCounter", Type: "counter", Value: counter})

	return metrics
}

func reportMetrics(metrics []model.Metric) error {
	for _, metric := range metrics {
		client := &http.Client{}
		endpoint := fmt.Sprintf("%s/%s/%s/%v", reportUrl, metric.Type, metric.Name, metric.Value)
		req, err := http.NewRequest(http.MethodPost, endpoint, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "text/plain")
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()
	}

	return nil
}
