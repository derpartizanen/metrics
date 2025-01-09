package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	Host string
}

func ConfigureServer() *ServerConfig {
	var host string

	flag.StringVar(&host, "a", "localhost:8080", "server host")
	flag.Parse()
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		host = envAddress
	}
	cfg := &ServerConfig{Host: host}

	return cfg
}
