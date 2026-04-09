package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService provides a thin abstraction over Redis for caching operations.
type CacheService struct {
	Client *redis.Client
}

// NewCacheService creates a CacheService backed by the given Redis client.
func NewCacheService(client *redis.Client) *CacheService {
	return &CacheService{Client: client}
}

// Get retrieves a cached value by key. Returns empty string and redis.Nil error on miss.
func (s *CacheService) Get(ctx context.Context, key string) (string, error) {
	return s.Client.Get(ctx, key).Result()
}

// Set stores a value in cache with the specified TTL.
func (s *CacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.Client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a key from the cache.
func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.Client.Del(ctx, key).Err()
}
