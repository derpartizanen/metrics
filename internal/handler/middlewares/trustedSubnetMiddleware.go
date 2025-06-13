package middlewares

import (
	"net"
	"net/http"
	"strings"

	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/network"
)

func (m *TrustedSubnetMiddleware) VerifySubnet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		remoteAddr := r.Header.Get("X-REAL-IP")
		if remoteAddr == "" {
			ra := strings.Split(r.RemoteAddr, ":")
			if len(ra) >= 2 {
				remoteAddr = ra[0]
			}
		}

		ip := net.ParseIP(remoteAddr)

		if ip == nil || !m.Subnet.Contains(ip) {
			if strings.Contains(contentType, "application/json") {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error": "your ip-address is not allowed"}`))
			} else {
				http.Error(w, "request from this ip-address was rejected", http.StatusForbidden)
			}
			logger.Log.Info("request rejected by ip address")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type TrustedSubnetMiddleware struct {
	Subnet *net.IPNet
}

func NewTrustedSubnetMiddleware(subnetStr string) *TrustedSubnetMiddleware {
	subnet := network.GetSubnetFromString(subnetStr)
	return &TrustedSubnetMiddleware{
		Subnet: subnet,
	}
}
