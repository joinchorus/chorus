package config

import (
	"os"
	"time"
)

// Config holds the application configuration.
type Config struct {
	Port         string
	Environment  string
	DataDir      string
	AdminToken   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load fetches configuration settings from environment variables or sensible defaults.
func Load() *Config {
	port := getEnv("PORT", "8085")
	env := getEnv("ENV", "development")
	dataDir := getEnv("DATA_DIR", "./data/repository")
	adminToken := getEnv("CHORUS_LICENSE_KEY", getEnv("CHORUS_ADMIN_TOKEN", getEnv("MODERATOR_API_KEY", "chorus-admin-secret-key-change-in-prod")))

	return &Config{
		Port:         port,
		Environment:  env,
		DataDir:      dataDir,
		AdminToken:   adminToken,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
