package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerAddr  string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		ServerAddr:  envOrDefault("SERVER_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
