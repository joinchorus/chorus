package geoip

import (
	"context"
	"net/http"
	"strings"

	"chorus/internal/domain"
)

// HeaderProvider reads the country code Cloudflare puts in CF-IPCountry. Callers
// must only use it for requests proven to come through Cloudflare.
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

// ResolveFromRequest returns the CF-IPCountry code, or "" when absent or unusable.
func (h *HeaderProvider) ResolveFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	if cfCountry := r.Header.Get("CF-IPCountry"); cfCountry != "" {
		code := strings.ToUpper(strings.TrimSpace(cfCountry))
		if err := domain.ValidateCountry(&code); err == nil && code != "" && code != "XX" && code != "T1" {
			return code
		}
	}

	return ""
}
