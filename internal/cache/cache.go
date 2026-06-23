package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss is returned when a key is not found in the cache.
var ErrCacheMiss = errors.New("cache: key not found")

// Cache wraps a Redis client with JSON serialization helpers.
type Cache struct {
	client *redis.Client
}

// NewCache creates a new Cache instance.
func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// Set serializes value as JSON and stores it in Redis with the given TTL.
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a key from Redis and deserializes the JSON into dest.
// Returns ErrCacheMiss if the key does not exist.
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return fmt.Errorf("cache get error: %w", err)
	}
	return json.Unmarshal(data, dest)
}

// Delete removes a key from the cache.
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
