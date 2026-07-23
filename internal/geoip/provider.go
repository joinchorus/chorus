package geoip

import "context"

// Provider abstracts IP geolocation providers (Cloudflare, MaxMind, Local/Dev).
type Provider interface {
	Name() string
	ResolveCountry(ctx context.Context, ipStr string) (string, error)
}
