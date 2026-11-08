package geoip_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"chorus/internal/geoip"
)

const secret = "test-edge-secret-value"

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
		name      string
		configure string // edge secret configured on the server
		remote    string
		headers   map[string]string
		want      string
	}{
		{"loopback peer gets the dev default", "", "127.0.0.1:12345", nil, "US"},
		{"private peer gets the dev default", "", "10.0.0.7:12345", nil, "US"},
		{"public peer without headers is unknown", "", "8.8.8.8:443", nil, "UN"},

		{"CF-IPCountry with the right secret is used", secret, "1.2.3.4:443",
			map[string]string{"CF-IPCountry": "de", geoip.EdgeSecretHeader: secret}, "DE"},
		{"CF-IPCountry with no secret header is ignored", secret, "1.2.3.4:443",
			map[string]string{"CF-IPCountry": "DE"}, "UN"},
		{"CF-IPCountry with a wrong secret is ignored", secret, "1.2.3.4:443",
			map[string]string{"CF-IPCountry": "DE", geoip.EdgeSecretHeader: "guess"}, "UN"},
		{"a guessed 'true' secret is ignored", secret, "1.2.3.4:443",
			map[string]string{"CF-IPCountry": "DE", geoip.EdgeSecretHeader: "true"}, "UN"},
		{"no secret configured ignores even a matching header", "", "1.2.3.4:443",
			map[string]string{"CF-IPCountry": "DE", geoip.EdgeSecretHeader: ""}, "UN"},
		{"Cloudflare unknown (XX) is unknown", secret, "1.2.3.4:443",
			map[string]string{"CF-IPCountry": "XX", geoip.EdgeSecretHeader: secret}, "UN"},
		{"Tor (T1) is unknown", secret, "1.2.3.4:443",
			map[string]string{"CF-IPCountry": "T1", geoip.EdgeSecretHeader: secret}, "UN"},
		{"X-Appengine-Country is never trusted, even with the secret", secret, "1.2.3.4:443",
			map[string]string{"X-Appengine-Country": "DE", geoip.EdgeSecretHeader: secret}, "UN"},

		{"spoofed X-Forwarded-For cannot pick the dev default", "", "8.8.8.8:443",
			map[string]string{"X-Forwarded-For": "127.0.0.1"}, "UN"},
		{"spoofed X-Real-IP cannot pick the dev default", "", "8.8.8.8:443",
			map[string]string{"X-Real-IP": "10.0.0.1"}, "UN"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := geoip.NewService("US", tc.configure)
			if got := svc.ResolveCountryFromRequest(ctx, request(tc.remote, tc.headers)); got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}

	t.Run("nil request is unknown", func(t *testing.T) {
		if got := geoip.NewService("US", secret).ResolveCountryFromRequest(ctx, nil); got != "UN" {
			t.Errorf("got %s, want UN", got)
		}
	})
}
