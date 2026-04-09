package config

import "os"

// Config holds all application configuration values.
type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis
	RedisURL string

	// JWT
	JWTSecret string
	JWTExpiry string // duration string, e.g. "24h"

	// Server
	Port string
}

// LoadConfig reads configuration from environment variables with sensible
// defaults so the application can start locally without any .env tooling.
func LoadConfig() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "go_fiber_api"),
		RedisURL:   getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:  getEnv("JWT_SECRET", "supersecretkey"),
		JWTExpiry:  getEnv("JWT_EXPIRY", "24h"),
		Port:       getEnv("PORT", "3000"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
