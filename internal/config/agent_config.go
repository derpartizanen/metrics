package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	ReportEndpoint string
	ReportInterval int
	PollInterval   int
	HashKey        string
}

type EnvParams struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	KEY            string `env:"KEY"`
}

func ConfigureAgent() *AgentConfig {
	var envParams EnvParams

	var reportEndpoint string
	var reportInterval int
	var pollInterval int
	var hashKey string

	flag.StringVar(&reportEndpoint, "a", "localhost:8080", "server host")
	flag.IntVar(&reportInterval, "r", 10, "report interval, seconds")
	flag.IntVar(&pollInterval, "p", 2, "poll interval, seconds")
	flag.StringVar(&hashKey, "k", "", "hash key")
	flag.Parse()

	err := env.Parse(&envParams)
	if err != nil {
		log.Fatal(err)
	}
	if envParams.Address != "" {
		reportEndpoint = envParams.Address
	}
	if envParams.ReportInterval != 0 {
		reportInterval = envParams.ReportInterval
	}
	if envParams.PollInterval != 0 {
		pollInterval = envParams.PollInterval
	}
	if envParams.KEY != "" {
		hashKey = envParams.KEY
	}
	cfg := &AgentConfig{
		ReportEndpoint: reportEndpoint, ReportInterval: reportInterval, PollInterval: pollInterval, HashKey: hashKey,
	}

	return cfg
}

func (cfg *AgentConfig) LogVars() {
	log.Printf("* reportEndpoint=%s\n", cfg.ReportEndpoint)
	log.Printf("* reportInterval=%d\n", cfg.ReportInterval)
	log.Printf("* pollInterval=%d\n", cfg.PollInterval)
}
