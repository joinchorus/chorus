package geoip

import (
	"context"
	"net/http"
	"strings"

	"chorus/internal/domain"
)

// HeaderProvider resolves country codes directly from trusted reverse proxy headers (Cloudflare CF-IPCountry, AppEngine, etc.).
type HeaderProvider struct{}

// NewHeaderProvider constructs a HeaderProvider.
func NewHeaderProvider() *HeaderProvider {
	return &HeaderProvider{}
}

func (h *HeaderProvider) Name() string {
	return "header"
}

func (h *HeaderProvider) ResolveCountry(ctx context.Context, ipStr string) (string, error) {
	// HeaderProvider resolves headers directly from HTTP request if context carries headers
	return "UN", nil
}

// ResolveFromRequest inspects request headers for edge proxy country codes.
func (h *HeaderProvider) ResolveFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	// 1. Cloudflare CF-IPCountry header
	if cfCountry := r.Header.Get("CF-IPCountry"); cfCountry != "" {
		code := strings.ToUpper(strings.TrimSpace(cfCountry))
		if err := domain.ValidateCountry(&code); err == nil && code != "" && code != "XX" && code != "T1" {
			return code
		}
	}

	// 2. GCP / AppEngine X-Appengine-Country header
	if gcpCountry := r.Header.Get("X-Appengine-Country"); gcpCountry != "" {
		code := strings.ToUpper(strings.TrimSpace(gcpCountry))
		if err := domain.ValidateCountry(&code); err == nil && code != "" {
			return code
		}
	}

	return ""
}
