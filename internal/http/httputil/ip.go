package httputil

import (
	"net/http"

	"chorus/internal/geoip"
)

// ExtractClientIP extracts the client IP address from HTTP request headers or RemoteAddr.
func ExtractClientIP(r *http.Request) string {
	return geoip.ExtractClientIP(r)
}
