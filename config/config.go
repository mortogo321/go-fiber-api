package config

import "os"

// Config holds all application configuration loaded from environment variables.
type Config struct {
	AppEnv     string
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	RedisURL   string
	JWTSecret  string
}

// Load reads configuration from environment variables with sensible defaults.
// In production set APP_ENV=production, DB_SSLMODE=require and a strong JWT_SECRET.
func Load() *Config {
	return &Config{
		AppEnv:     getEnv("APP_ENV", "development"),
		Port:       getEnv("PORT", "3000"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "gofiber"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		RedisURL:   getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:  getEnv("JWT_SECRET", "changeme"),
	}
}

// IsProduction reports whether the app runs in production mode.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
