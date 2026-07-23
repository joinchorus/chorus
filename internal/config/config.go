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
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load fetches configuration settings from environment variables or sensible defaults.
func Load() *Config {
	port := getEnv("PORT", "8085")
	env := getEnv("ENV", "development")
	dataDir := getEnv("DATA_DIR", "./data/repository")

	return &Config{
		Port:         port,
		Environment:  env,
		DataDir:      dataDir,
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
