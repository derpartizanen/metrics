package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	Address          string `env:"ADDRESS" json:"address"`
	ReportInterval   int    `env:"REPORT_INTERVAL" json:"report_interval"`
	ReportRetryCount int    `env:"REPORT_RETRY_COUNT" json:"report_retry_count"`
	PollInterval     int    `env:"POLL_INTERVAL" json:"poll_interval"`
	HashKey          string `env:"KEY" json:"key"`
	RateLimit        int    `env:"RATE_LIMIT" json:"rate_limit"`
	CryptoKey        string `env:"CRYPTO_KEY" json:"crypto_key"`
}

func ConfigureAgent() *AgentConfig {
	config := &AgentConfig{}

	flag.StringVar(&config.Address, "a", "localhost:8080", "server host")
	flag.IntVar(&config.ReportInterval, "r", 10, "report interval, seconds")
	flag.IntVar(&config.ReportRetryCount, "c", 3, "report retry count")
	flag.IntVar(&config.PollInterval, "p", 2, "poll interval, seconds")
	flag.StringVar(&config.HashKey, "k", "", "hash key")
	flag.IntVar(&config.RateLimit, "l", 1, "rate limit")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "crypto key")
	var configPath string
	flag.StringVar(&configPath, "config", "", "config file")
	flag.Parse()

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configPath = envConfig
	}

	if configPath != "" {
		err := config.loadAgentConfigFile(configPath)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to load config file '%s': %w", configPath, err))
		}
	}

	if err := env.Parse(config); err != nil {
		log.Fatal(fmt.Errorf("failed to parse config: %w", err))
	}

	return config
}

func (cfg *AgentConfig) LogVars() {
	log.Printf("* reportEndpoint=%s\n", cfg.Address)
	log.Printf("* reportInterval=%d\n", cfg.ReportInterval)
	log.Printf("* reportRetryCount=%d\n", cfg.ReportRetryCount)
	log.Printf("* pollInterval=%d\n", cfg.PollInterval)
	log.Printf("* rateLimit=%d\n", cfg.RateLimit)
}

func (cfg *AgentConfig) loadAgentConfigFile(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read file by path '%s': %w", configPath, err)
	}

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data '%s': %w", string(data), err)
	}

	return nil
}
