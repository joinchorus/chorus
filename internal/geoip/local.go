package geoip

import (
	"context"
	"net"
	"strings"
)

// LocalProvider resolves loopback and private IP addresses during development.
type LocalProvider struct {
	defaultCountry string
}

// NewLocalProvider constructs a LocalProvider.
func NewLocalProvider(defaultCountry string) *LocalProvider {
	defaultCountry = strings.ToUpper(strings.TrimSpace(defaultCountry))
	if defaultCountry == "" {
		defaultCountry = "US"
	}
	return &LocalProvider{
		defaultCountry: defaultCountry,
	}
}

func (l *LocalProvider) Name() string {
	return "local"
}

func (l *LocalProvider) ResolveCountry(ctx context.Context, ipStr string) (string, error) {
	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "" {
		return "UN", nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ipStr == "localhost" {
		return l.defaultCountry, nil
	}

	return "UN", nil
}
