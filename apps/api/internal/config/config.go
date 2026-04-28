package config

import (
	"fmt"
	"os"
)

// Config holds process configuration loaded from the environment.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	StorageDir  string
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://finance:finance@localhost:5432/finance_insights?sslmode=disable"),
		StorageDir:  getenv("STORAGE_DIR", "storage"),
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("HTTP_ADDR must not be empty")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must be set")
	}
	if cfg.StorageDir == "" {
		return Config{}, fmt.Errorf("STORAGE_DIR must not be empty")
	}
	return cfg, nil
}

func getenv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
