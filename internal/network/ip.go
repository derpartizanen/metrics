package network

import (
	"errors"
	"fmt"
	"net"
)

func GetIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to identify IP addresses: %w", err)
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.To4(), nil
		}
	}
	return nil, errors.New("ip address is not found")
}

func GetSubnetFromString(subnet string) *net.IPNet {
	if subnet == "" {
		return nil
	}
	_, s, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil
	}

	return s
}
