package config

import (
	"os"
	"time"
)

// Config holds the application configuration.
type Config struct {
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load fetches configuration settings from environment variables or sensible defaults.
func Load() *Config {
	port := getEnv("PORT", "8080")
	env := getEnv("ENV", "development")

	return &Config{
		Port:         port,
		Environment:  env,
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
