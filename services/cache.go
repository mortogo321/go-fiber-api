package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService wraps a Redis client with convenience methods for cache operations.
type CacheService struct {
	client *redis.Client
}

// NewCacheService creates a new CacheService backed by the given Redis client.
func NewCacheService(client *redis.Client) *CacheService {
	return &CacheService{client: client}
}

// Get retrieves a cached value by key. Returns an empty string and an error on miss.
func (s *CacheService) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

// Set stores a value in the cache with the given TTL.
func (s *CacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a specific key from the cache.
func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

// InvalidatePattern removes all keys matching the given glob pattern.
// Uses SCAN internally to avoid blocking the Redis server on large key spaces.
func (s *CacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	iter := s.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := s.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
