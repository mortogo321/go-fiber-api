package config

import "os"

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	RedisURL   string
	JWTSecret  string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "3000"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "gofiber"),
		RedisURL:   getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:  getEnv("JWT_SECRET", "changeme"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
