package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/redis/go-redis/v9"
)

type LinkCache struct {
	rdb     *redis.Client
	linkTTL time.Duration
}

const linkKeyPrefix = "link:"

func NewRedisRepo(rdb *redis.Client, ttl time.Duration) *LinkCache {
	return &LinkCache{rdb: rdb, linkTTL: ttl}
}

func (c *LinkCache) Get(ctx context.Context, shortCode string) (string, error) {
	key := getLinkKey(shortCode)

	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", domain.ErrCacheMiss
		}
		return "", fmt.Errorf("failed to get link %s from redis: %w", shortCode, err)
	}
	return val, nil
}

func (c *LinkCache) Set(ctx context.Context, shortCode, originalURL string) error {
	key := getLinkKey(shortCode)

	err := c.rdb.Set(ctx, key, originalURL, c.linkTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set link %s to redis: %w", shortCode, err)
	}

	return nil
}

func getLinkKey(shortCode string) string {
	return linkKeyPrefix + shortCode
}
