package redis

import (
	"context"
	"fmt"

	"github.com/ismailtemuroglu/discord/internal/platform/config"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	cache *redis.Client
}

func NewCache(ctx context.Context, cfg *config.Config) (*Cache, error) {
	cache := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := cache.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Cache{cache: cache}, nil
}

func (c *Cache) Close() error {
	return c.cache.Close()
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.cache.Ping(ctx).Err()
}

func (c *Cache) GetClient() *redis.Client {
	return c.cache
}	