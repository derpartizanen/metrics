package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Host          string `env:"ADDRESS" json:"address"`
	StoragePath   string `env:"STORAGE_PATH" json:"store_file"`
	StoreInterval int64  `env:"STORE_INTERVAL" json:"store_interval"`
	Restore       bool   `env:"RESTORE" json:"restore"`
	Loglevel      string `env:"LOG_LEVEL" json:"log_level"`
	DatabaseDSN   string `env:"DATABASE_DSN" json:"database_dsn"`
	Key           string `env:"KEY" json:"key"`
	CryptoKey     string `env:"CRYPTO_KEY" json:"crypto_key"`
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	GrpcAddress   string `env:"GRPC_ADDRESS" json:"grpc_address"`
}

func ConfigureServer() *ServerConfig {
	config := &ServerConfig{}

	flag.StringVar(&config.Host, "a", "localhost:8080", "server host")
	flag.StringVar(&config.StoragePath, "f", "/tmp/metrics-storage.json", "path to file to store metrics")
	flag.Int64Var(&config.StoreInterval, "i", 300, "interval of storing metrics")
	flag.BoolVar(&config.Restore, "r", true, "load metrics from file")
	flag.StringVar(&config.DatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&config.Key, "k", "", "hash key")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "crypto key")
	flag.StringVar(&config.TrustedSubnet, "t", "", "trusted subnet (CIDR)")
	flag.StringVar(&config.Loglevel, "l", "DEBUG", "log level")
	flag.StringVar(&config.GrpcAddress, "grpc-address", "localhost:9090", "grpc address")
	var configPath string
	flag.StringVar(&configPath, "config", "", "config file")
	flag.Parse()

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configPath = envConfig
	}

	if configPath != "" {
		err := config.loadServerConfigFile(configPath)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to load config file '%s': %w", configPath, err))
		}
	}

	if err := env.Parse(config); err != nil {
		log.Fatal(fmt.Errorf("failed to parse config: %w", err))
	}

	return config
}

func (cfg *ServerConfig) loadServerConfigFile(configPath string) error {
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

func (cfg *ServerConfig) LogVars() {
	log.Printf("* Address=%s\n", cfg.Host)
	log.Printf("* Grpc Address=%s\n", cfg.GrpcAddress)
	log.Printf("* StorePath=%s\n", cfg.StoragePath)
	log.Printf("* StoreInterval=%d\n", cfg.StoreInterval)
	log.Printf("* Restore=%t\n", cfg.Restore)
	log.Printf("* CryptoKey=%s\n", cfg.CryptoKey)
	log.Printf("* TrustedSubnet=%s\n", cfg.TrustedSubnet)
}
