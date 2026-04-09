package database

import (
	"context"
	"log"

	"github.com/mor-tesla/go-fiber-api/config"
	"github.com/redis/go-redis/v9"
)

// ConnectRedis initializes and verifies a Redis client connection.
func ConnectRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	log.Println("connected to Redis")
	return rdb
}
