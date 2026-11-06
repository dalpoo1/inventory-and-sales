package config

import "os"

// Config holds runtime configuration, loaded from the environment.
type Config struct {
	Port        string
	DatabaseURL string
}

// Load reads configuration from environment variables with sane defaults.
func Load() Config {
	return Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://stockflow:stockflow@localhost:5432/stockflow?sslmode=disable"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
