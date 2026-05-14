package config

import (
	"os"
	"strconv"
)

// Config holds runtime configuration loaded from the environment.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTExpiryH  int
}

// Load reads configuration from environment variables with sensible defaults for local dev.
func Load() Config {
	expiry := 72
	if v := os.Getenv("JWT_EXPIRY_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			expiry = n
		}
	}
	return Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/shopapi?sslmode=disable"),
		JWTSecret:   getenv("JWT_SECRET", "change-me-in-production-use-long-random-secret"),
		JWTExpiryH:  expiry,
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
