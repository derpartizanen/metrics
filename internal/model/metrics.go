package model

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

var (
	GaugeMetrics = []string{
		"Alloc",
		"TotalAlloc",
		"Sys",
		"Lookups",
		"Mallocs",
		"Frees",
		"HeapAlloc",
		"HeapSys",
		"HeapIdle",
		"HeapInuse",
		"HeapReleased",
		"HeapObjects",
		"StackInuse",
		"StackSys",
		"MSpanInuse",
		"MSpanSys",
		"MCacheInuse",
		"MCacheSys",
		"BuckHashSys",
		"GCSys",
		"OtherSys",
		"NextGC",
		"LastGC",
		"PauseTotalNs",
		"GCCPUFraction",
		"NumForcedGC",
		"NumGC",
	}
)
