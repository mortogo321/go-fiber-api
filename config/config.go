package config

import "os"

// Config holds all application configuration values loaded from environment variables.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	RedisURL   string
	JWTSecret  string
	Port       string
}

// LoadConfig reads configuration from environment variables with sensible defaults.
func LoadConfig() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "fiber_api"),
		RedisURL:   getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:  getEnv("JWT_SECRET", "default-secret-change-me"),
		Port:       getEnv("PORT", "3000"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
