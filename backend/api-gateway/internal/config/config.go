package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	PolicyPath  string
}

// Load reads runtime configuration from environment variables with local defaults.
func Load() Config {
	return Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://secure:secure@localhost:5432/secure_deploy?sslmode=disable"),
		PolicyPath:  getEnv("POLICY_PATH", "policies"),
	}
}

// getEnv returns an environment variable when present, otherwise it falls back to the provided default.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
