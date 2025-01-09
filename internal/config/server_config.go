package config

import (
	"flag"
)

type ServerConfig struct {
	Host string
}

func ConfigureServer() *ServerConfig {
	var host string

	flag.StringVar(&host, "a", "localhost:8080", "server host")
	flag.Parse()
	cfg := &ServerConfig{Host: host}

	return cfg
}
