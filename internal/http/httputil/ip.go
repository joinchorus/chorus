package httputil

import (
	"net/http"

	"chorus/internal/geoip"
)

// ExtractClientIP extracts the client IP address from HTTP request headers or RemoteAddr.
func ExtractClientIP(r *http.Request) string {
	return geoip.ExtractClientIP(r)
}

// ResolveCountryFromIP resolves the country code based on the client IP address.
func ResolveCountryFromIP(ipStr string) string {
	provider := geoip.NewLocalProvider("US")
	code, err := provider.ResolveCountry(nil, ipStr)
	if err != nil {
		return "UN"
	}
	return code
}
