package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/derpartizanen/metrics/internal/config"
	"github.com/derpartizanen/metrics/internal/model"
)

var cfg *config.AgentConfig
var counter int64

func main() {
	var metrics []model.Metrics

	cfg = config.ConfigureAgent()
	log.Println("Starting agent")
	cfg.LogVars()

	go func() {
		for {
			metrics = updateMetrics()
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		}
	}()

	for {
		time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
		reportMetrics(metrics)
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

func reportMetrics(metrics []model.Metrics) {
	reportURL := fmt.Sprintf("http://%s/update/", cfg.ReportEndpoint)
	client := &http.Client{}
	for _, metric := range metrics {
		jsonStr, err := json.Marshal(metric)
		log.Printf("report metric %s with body %s", metric.ID, jsonStr)
		if err != nil {
			log.Print(err)
			continue
		}
		req, err := http.NewRequest(http.MethodPost, reportURL, bytes.NewBuffer(jsonStr))
		if err != nil {
			log.Print(err)
			continue
		}
		req.Header.Add("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			log.Print("request error: ", err)
			continue
		}
		res.Body.Close()
	}
}
