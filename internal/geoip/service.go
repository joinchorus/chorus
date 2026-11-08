package geoip

import (
	"context"
	"crypto/subtle"
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
	// edgeSecret proves a request came through our Cloudflare zone. The zone
	// adds it as the EdgeSecretHeader request header; the country header is
	// believed only when the value matches. Empty disables edge headers.
	edgeSecret string
}

// EdgeSecretHeader carries the shared secret that Cloudflare adds (Transform
// Rule) to every request for the app hostname.
const EdgeSecretHeader = "X-Chorus-Edge"

// NewService constructs a concrete GeoIP Service with provider chain.
// defaultDevCountry is returned for loopback/private peers (local development).
// edgeSecret is the value Cloudflare sends in EdgeSecretHeader; pass "" to
// ignore CF-IPCountry entirely (no trusted edge).
func NewService(defaultDevCountry string, edgeSecret string) *Service {
	if defaultDevCountry == "" {
		defaultDevCountry = "TR"
	}
	localP := NewLocalProvider(defaultDevCountry)
	headerP := NewHeaderProvider()

	return &Service{
		providers: []Provider{headerP, localP},
		headerP:   headerP,
		localP:    localP,

		edgeSecret: edgeSecret,
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

	// 1. CF-IPCountry, only when the request proves it passed through our edge.
	if s.fromTrustedEdge(r) {
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

// fromTrustedEdge reports whether the request carries the correct edge secret.
// A direct hit on the origin (e.g. the *.hosted.app URL) cannot know it.
func (s *Service) fromTrustedEdge(r *http.Request) bool {
	if s.edgeSecret == "" {
		return false
	}
	got := r.Header.Get(EdgeSecretHeader)
	return subtle.ConstantTimeCompare([]byte(got), []byte(s.edgeSecret)) == 1
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
