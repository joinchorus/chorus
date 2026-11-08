package config

import (
	"errors"
	"os"
	"time"
)

// DefaultAdminToken is the development fallback used when no admin token is
// configured. It is public (it lives in this repository), so Validate rejects
// it in production.
const DefaultAdminToken = "chorus-admin-secret-key-change-in-prod"

// ErrInsecureAdminToken is returned by Validate when production runs without a
// real admin token.
var ErrInsecureAdminToken = errors.New(
	"ENV=production requires an admin token: set CHORUS_ADMIN_TOKEN (or CHORUS_LICENSE_KEY / MODERATOR_API_KEY)")

// Config holds the application configuration.
type Config struct {
	Port        string
	Environment string
	DataDir     string
	AdminToken  string
	// EdgeSecret is the shared secret Cloudflare adds as X-Chorus-Edge. When set,
	// CF-IPCountry is believed for requests that carry it. Empty disables it.
	EdgeSecret   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load fetches configuration settings from environment variables or sensible defaults.
func Load() *Config {
	port := getEnv("PORT", "8085")
	env := getEnv("ENV", "development")
	dataDir := getEnv("DATA_DIR", "./data/repository")
	adminToken := getEnv("CHORUS_LICENSE_KEY", getEnv("CHORUS_ADMIN_TOKEN", getEnv("MODERATOR_API_KEY", DefaultAdminToken)))

	return &Config{
		Port:         port,
		Environment:  env,
		DataDir:      dataDir,
		AdminToken:   adminToken,
		EdgeSecret:   getEnv("CHORUS_EDGE_SECRET", ""),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// Validate reports configuration that is unsafe to run with. In production the
// admin token must be set and must not be the public placeholder.
func (c *Config) Validate() error {
	if c.Environment == "production" && (c.AdminToken == "" || c.AdminToken == DefaultAdminToken) {
		return ErrInsecureAdminToken
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
