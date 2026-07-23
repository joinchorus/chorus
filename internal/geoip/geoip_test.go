package geoip_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"chorus/internal/geoip"
)

func TestGeoIPService(t *testing.T) {
	svc := geoip.NewService("US")
	ctx := context.Background()

	t.Run("resolve loopback IP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"

		code := svc.ResolveCountryFromRequest(ctx, req)
		if code != "US" {
			t.Errorf("expected US for loopback IP, got %s", code)
		}
	})

	t.Run("resolve CF-IPCountry header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("CF-IPCountry", "DE")
		req.RemoteAddr = "1.2.3.4:12345"

		code := svc.ResolveCountryFromRequest(ctx, req)
		if code != "DE" {
			t.Errorf("expected DE from CF-IPCountry header, got %s", code)
		}
	})

	t.Run("resolve X-Forwarded-For loopback IP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.100, 10.0.0.1")

		code := svc.ResolveCountryFromRequest(ctx, req)
		if code != "US" {
			t.Errorf("expected default US for private network IP, got %s", code)
		}
	})
}
