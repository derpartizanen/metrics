package config

import (
	"flag"
	"log"
)

type AgentConfig struct {
	ReportEndpoint string
	ReportInterval int
	PollInterval   int
}

func ConfigureAgent() *AgentConfig {
	var reportEndpoint string
	var reportInterval int
	var pollInterval int

	flag.StringVar(&reportEndpoint, "a", "localhost:8080", "server host")
	flag.IntVar(&reportInterval, "r", 10, "report interval, seconds")
	flag.IntVar(&pollInterval, "p", 2, "poll interval, seconds")
	flag.Parse()
	cfg := &AgentConfig{ReportEndpoint: reportEndpoint, ReportInterval: reportInterval, PollInterval: pollInterval}

	return cfg
}

func (cfg *AgentConfig) LogVars() {
	log.Printf("* reportEndpoint=%s\n", cfg.ReportEndpoint)
	log.Printf("* reportInterval=%d\n", cfg.ReportInterval)
	log.Printf("* pollInterval=%d\n", cfg.PollInterval)
}
