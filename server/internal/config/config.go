package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	ServerAddr          string
	GRPCAddr            string
	DatabaseURL         string
	MigrationsDir       string
	ChallengeTTL        time.Duration
	SessionTTL          time.Duration
	SecurityCenterAddr  string
}

func Load() (Config, error) {
	challengeTTL, err := parseDurationEnv("CHALLENGE_TTL", 60*time.Second)
	if err != nil {
		return Config{}, err
	}

	sessionTTL, err := parseDurationEnv("SESSION_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		ServerAddr:         envOrDefault("SERVER_ADDR", ":8080"),
		GRPCAddr:           envOrDefault("GRPC_ADDR", ":50051"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		MigrationsDir:      envOrDefault("MIGRATIONS_DIR", "migrations"),
		ChallengeTTL:       challengeTTL,
		SessionTTL:         sessionTTL,
		SecurityCenterAddr: os.Getenv("SECURITY_CENTER_ADDR"),
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

func parseDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return duration, nil
}
