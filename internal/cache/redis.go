package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/gupta/leetcode-judge/internal/common"
	"github.com/gupta/leetcode-judge/internal/config"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a Redis connection, pings it, and returns the client.
func NewRedisClient(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	common.Logger.Info("Redis connected successfully")
	return client, nil
}
