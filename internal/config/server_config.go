package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Host          string `env:"ADDRESS"`
	StoragePath   string `env:"STORAGE_PATH"`
	StoreInterval int64  `env:"STORE_INTERVAL"`
	Restore       bool   `env:"RESTORE"`
	Loglevel      string `env:"LOG_LEVEL"`
}

func ConfigureServer() *ServerConfig {
	config := &ServerConfig{}
	env.Parse(config)

	flagConfig := parseServerFlags()
	if config.Host == "" {
		config.Host = flagConfig.Host
	}
	if config.StoreInterval == 0 {
		config.StoreInterval = flagConfig.StoreInterval
	}
	if config.StoragePath == "" {
		config.StoragePath = flagConfig.StoragePath
	}
	if !config.Restore {
		config.Restore = flagConfig.Restore
	}
	if config.Loglevel == "" {
		config.Loglevel = flagConfig.Loglevel
	}

	return config
}

func parseServerFlags() *ServerConfig {
	config := &ServerConfig{}
	flag.StringVar(&config.Host, "a", "localhost:8080", "server host")
	flag.StringVar(&config.StoragePath, "f", "/tmp/metrics-storage.json", "path to file to store metrics")
	flag.Int64Var(&config.StoreInterval, "i", 300, "interval of storing metrics")
	flag.BoolVar(&config.Restore, "r", true, "load metrics from file")
	flag.StringVar(&config.Loglevel, "l", "DEBUG", "log level")
	flag.Parse()

	return config
}
