package geoip

import (
	"context"
	"net"
	"net/http"
	"strings"

	"chorus/internal/domain"
)

// Service resolves ISO 3166-1 alpha-2 country codes from client IP addresses and headers.
type Service struct {
	providers []Provider
	headerP   *HeaderProvider
	localP    *LocalProvider
}

// NewService constructs a concrete GeoIP Service with provider chain.
func NewService(defaultDevCountry string) *Service {
	if defaultDevCountry == "" {
		defaultDevCountry = "TR"
	}
	localP := NewLocalProvider(defaultDevCountry)
	headerP := NewHeaderProvider()

	return &Service{
		providers: []Provider{headerP, localP},
		headerP:   headerP,
		localP:    localP,
	}
}

// ResolveCountryFromRequest resolves ISO 3166-1 alpha-2 country from HTTP request headers & client IP.
func (s *Service) ResolveCountryFromRequest(ctx context.Context, r *http.Request) string {
	if r == nil {
		return "UN"
	}

	// 1. Try edge headers first (CF-IPCountry, etc.)
	if code := s.headerP.ResolveFromRequest(r); code != "" {
		return code
	}

	// 2. Extract client IP address
	clientIP := ExtractClientIP(r)

	// 3. Fallback to Local/IP provider
	code, err := s.localP.ResolveCountry(ctx, clientIP)
	if err == nil && code != "" {
		if err := domain.ValidateCountry(&code); err == nil {
			return code
		}
	}

	return "UN"
}

// ExtractClientIP extracts the client IP address from HTTP request headers or RemoteAddr.
func ExtractClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

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
