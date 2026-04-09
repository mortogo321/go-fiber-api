package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService wraps a Redis client and exposes simple cache operations used
// throughout the application.
type CacheService struct {
	client *redis.Client
}

// NewCacheService creates a CacheService backed by the given Redis client.
func NewCacheService(client *redis.Client) *CacheService {
	return &CacheService{client: client}
}

// Get retrieves a cached value by key. Returns an empty string and a non-nil
// error on a cache miss.
func (s *CacheService) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.client.Get(ctx, key).Result()
}

// Set stores a key-value pair with the specified TTL.
func (s *CacheService) Set(key string, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a single key from the cache.
func (s *CacheService) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.client.Del(ctx, key).Err()
}

// DeleteByPattern removes all keys matching the given glob pattern. This uses
// SCAN internally to avoid blocking Redis with a KEYS command.
func (s *CacheService) DeleteByPattern(pattern string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cursor uint64
	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := s.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
