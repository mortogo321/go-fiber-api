package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mortogo321/go-fiber-api/config"
	"github.com/redis/go-redis/v9"
)

// RedisClient is the global Redis client instance.
var RedisClient *redis.Client

// InitRedis parses the Redis URL, creates a client, and verifies connectivity.
func InitRedis(cfg *config.Config) error {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("failed to parse redis URL: %w", err)
	}

	RedisClient = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}
