package geoip_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"chorus/internal/geoip"
)

func request(remoteAddr string, headers map[string]string) *http.Request {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}

func TestResolveCountryFromRequest(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		trust   bool
		remote  string
		headers map[string]string
		want    string
	}{
		{"loopback peer gets the dev default", false, "127.0.0.1:12345", nil, "US"},
		{"private peer gets the dev default", false, "10.0.0.7:12345", nil, "US"},
		{"public peer without headers is unknown", false, "8.8.8.8:443", nil, "UN"},
		{"trusted CF-IPCountry is used", true, "1.2.3.4:443", map[string]string{"CF-IPCountry": "DE"}, "DE"},
		{"trusted X-Appengine-Country is used", true, "1.2.3.4:443", map[string]string{"X-Appengine-Country": "fr"}, "FR"},
		{"trusted but unusable CF value falls back to unknown", true, "1.2.3.4:443", map[string]string{"CF-IPCountry": "XX"}, "UN"},
		{"untrusted CF-IPCountry is ignored", false, "1.2.3.4:443", map[string]string{"CF-IPCountry": "DE"}, "UN"},
		{"untrusted X-Appengine-Country is ignored", false, "1.2.3.4:443", map[string]string{"X-Appengine-Country": "DE"}, "UN"},
		{"spoofed X-Forwarded-For cannot pick the dev default", false, "8.8.8.8:443", map[string]string{"X-Forwarded-For": "127.0.0.1"}, "UN"},
		{"spoofed X-Real-IP cannot pick the dev default", false, "8.8.8.8:443", map[string]string{"X-Real-IP": "10.0.0.1"}, "UN"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := geoip.NewService("US", tc.trust)
			if got := svc.ResolveCountryFromRequest(ctx, request(tc.remote, tc.headers)); got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}

	t.Run("nil request is unknown", func(t *testing.T) {
		if got := geoip.NewService("US", true).ResolveCountryFromRequest(ctx, nil); got != "UN" {
			t.Errorf("got %s, want UN", got)
		}
	})
}
