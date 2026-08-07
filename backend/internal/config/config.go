package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	DatabaseURL   string
	EncryptionKey string
	AllowedOrigin string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://myagents:myagents@localhost:5432/myagents?sslmode=disable"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
	}

	if len(cfg.EncryptionKey) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes (got %d) — used as an AES-256 key to encrypt model API keys at rest", len(cfg.EncryptionKey))
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
