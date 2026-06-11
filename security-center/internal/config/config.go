package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerAddr    string
	GRPCAddr      string
	DatabaseURL   string
	MigrationsDir string
	AdminToken    string
}

func Load() (Config, error) {
	cfg := Config{
		ServerAddr:    envOrDefault("SECURITY_SERVER_ADDR", ":8081"),
		GRPCAddr:      envOrDefault("SECURITY_GRPC_ADDR", ":50052"),
		DatabaseURL:   os.Getenv("SECURITY_DATABASE_URL"),
		MigrationsDir: envOrDefault("SECURITY_MIGRATIONS_DIR", "migrations"),
		AdminToken:    os.Getenv("SECURITY_ADMIN_TOKEN"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("SECURITY_DATABASE_URL is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
