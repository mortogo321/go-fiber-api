package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mortogo321/go-fiber-api/config"
	"github.com/redis/go-redis/v9"
)

// InitRedis creates a go-redis client and verifies connectivity with a PING.
func InitRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisURL,
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	log.Println("redis connected")
	return rdb, nil
}
