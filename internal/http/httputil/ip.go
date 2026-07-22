package httputil

import (
	"net"
	"net/http"
	"strings"
)

// ExtractClientIP extracts the client IP address from HTTP request headers or RemoteAddr.
func ExtractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// ResolveCountryFromIP resolves the country code based on the client IP address.
// In production, this can call GeoIP database/service. For local/development, resolves loopback to "US".
func ResolveCountryFromIP(ipStr string) string {
	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "127.0.0.1" || ipStr == "::1" || ipStr == "localhost" || strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") {
		return "US" // Default development country code
	}
	// Fallback ISO country code for unknown IP
	return "UN"
}
