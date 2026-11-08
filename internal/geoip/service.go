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
	// trustEdgeHeaders controls whether CF-IPCountry / X-Appengine-Country are
	// believed. Any client can send them, so this must only be true when every
	// request passes through a proxy that overwrites them (e.g. Cloudflare).
	trustEdgeHeaders bool
}

// NewService constructs a concrete GeoIP Service with provider chain.
// defaultDevCountry is returned for loopback/private peers (local development).
// trustEdgeHeaders enables the edge country headers; leave it false unless the
// app is only reachable through a proxy that sets them.
func NewService(defaultDevCountry string, trustEdgeHeaders bool) *Service {
	if defaultDevCountry == "" {
		defaultDevCountry = "TR"
	}
	localP := NewLocalProvider(defaultDevCountry)
	headerP := NewHeaderProvider()

	return &Service{
		providers: []Provider{headerP, localP},
		headerP:   headerP,
		localP:    localP,

		trustEdgeHeaders: trustEdgeHeaders,
	}
}

// ResolveCountryFromRequest resolves ISO 3166-1 alpha-2 country for a request.
// It returns "UN" when the country cannot be determined reliably.
//
// X-Forwarded-For and X-Real-IP are deliberately not used here: the client
// controls them, so they could pick any country (or the dev default).
func (s *Service) ResolveCountryFromRequest(ctx context.Context, r *http.Request) string {
	if r == nil {
		return "UN"
	}

	// 1. Edge headers, only when explicitly trusted.
	if s.trustEdgeHeaders {
		if code := s.headerP.ResolveFromRequest(r); code != "" {
			return code
		}
	}

	// 2. The direct TCP peer: only loopback/private peers resolve (development).
	code, err := s.localP.ResolveCountry(ctx, peerIP(r))
	if err == nil && code != "" {
		if err := domain.ValidateCountry(&code); err == nil {
			return code
		}
	}

	return "UN"
}

// peerIP returns the host of the direct TCP peer (RemoteAddr), ignoring headers.
func peerIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
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
