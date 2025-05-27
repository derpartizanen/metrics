package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	ReportEndpoint   string
	ReportInterval   int
	PollInterval     int
	HashKey          string
	RateLimit        int
	ReportRetryCount int
	CryptoKey        string
}

type EnvParams struct {
	Address          string `env:"ADDRESS"`
	ReportInterval   int    `env:"REPORT_INTERVAL"`
	ReportRetryCount int    `env:"REPORT_RETRY_COUNT"`
	PollInterval     int    `env:"POLL_INTERVAL"`
	KEY              string `env:"KEY"`
	RateLimit        int    `env:"RATE_LIMIT"`
	CryptoKey        string `env:"CRYPTO_KEY"`
}

func ConfigureAgent() *AgentConfig {
	var envParams EnvParams

	var reportEndpoint string
	var reportInterval int
	var reportRetryCount int
	var pollInterval int
	var hashKey string
	var rateLimit int
	var cryptoKey string

	flag.StringVar(&reportEndpoint, "a", "localhost:8080", "server host")
	flag.IntVar(&reportInterval, "r", 10, "report interval, seconds")
	flag.IntVar(&reportRetryCount, "c", 3, "report retry count")
	flag.IntVar(&pollInterval, "p", 2, "poll interval, seconds")
	flag.StringVar(&hashKey, "k", "", "hash key")
	flag.IntVar(&rateLimit, "l", 1, "rate limit")
	flag.StringVar(&cryptoKey, "crypto-key", "", "crypto key")
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
	if envParams.ReportRetryCount != 0 {
		reportInterval = envParams.ReportRetryCount
	}
	if envParams.PollInterval != 0 {
		pollInterval = envParams.PollInterval
	}
	if envParams.KEY != "" {
		hashKey = envParams.KEY
	}
	if envParams.RateLimit != 0 {
		rateLimit = envParams.RateLimit
	}
	if envParams.CryptoKey != "" {
		cryptoKey = envParams.CryptoKey
	}
	cfg := &AgentConfig{
		ReportEndpoint:   reportEndpoint,
		ReportInterval:   reportInterval,
		ReportRetryCount: reportRetryCount,
		PollInterval:     pollInterval,
		HashKey:          hashKey,
		RateLimit:        rateLimit,
		CryptoKey:        cryptoKey,
	}

	return cfg
}

func (cfg *AgentConfig) LogVars() {
	log.Printf("* reportEndpoint=%s\n", cfg.ReportEndpoint)
	log.Printf("* reportInterval=%d\n", cfg.ReportInterval)
	log.Printf("* reportRetryCount=%d\n", cfg.ReportRetryCount)
	log.Printf("* pollInterval=%d\n", cfg.PollInterval)
	log.Printf("* rateLimit=%d\n", cfg.RateLimit)
}
