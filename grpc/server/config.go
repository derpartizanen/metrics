package server

import (
	"net"

	"github.com/derpartizanen/metrics/internal/storage"
)

var config Config

type Config struct {
	PrivateKey    []byte
	ServerAddr    string
	SecretKey     string
	Service       *storage.Storage
	TrustedSubnet *net.IPNet
}
