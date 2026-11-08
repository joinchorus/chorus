package config

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		token   string
		wantErr bool
	}{
		{name: "production with placeholder token is rejected", env: "production", token: DefaultAdminToken, wantErr: true},
		{name: "production with empty token is rejected", env: "production", token: "", wantErr: true},
		{name: "production with real token is accepted", env: "production", token: "a-real-secret", wantErr: false},
		{name: "development may use the placeholder", env: "development", token: DefaultAdminToken, wantErr: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{Environment: tc.env, AdminToken: tc.token}
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestLoadFallsBackToPlaceholder(t *testing.T) {
	for _, key := range []string{"CHORUS_LICENSE_KEY", "CHORUS_ADMIN_TOKEN", "MODERATOR_API_KEY"} {
		t.Setenv(key, "")
	}
	if got := Load().AdminToken; got != DefaultAdminToken {
		t.Fatalf("expected placeholder token, got %q", got)
	}
}

func TestLoadEdgeSecret(t *testing.T) {
	t.Run("empty by default", func(t *testing.T) {
		t.Setenv("CHORUS_EDGE_SECRET", "")
		if got := Load().EdgeSecret; got != "" {
			t.Errorf("EdgeSecret must default to empty, got %q", got)
		}
	})
	t.Run("read from the environment", func(t *testing.T) {
		t.Setenv("CHORUS_EDGE_SECRET", "abc123")
		if got := Load().EdgeSecret; got != "abc123" {
			t.Errorf("got %q, want abc123", got)
		}
	})
}
