// Package redis provides Redis caching for resolved link URLs.
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/redis/go-redis/v9"
)

// LinkCache implements usecase.LinkCache backed by Redis.
type LinkCache struct {
	rdb     *redis.Client
	linkTTL time.Duration
}

const linkKeyPrefix = "link:"

// NewRedisRepo creates a new LinkCache with the provided Redis client and TTL duration.
func NewRedisRepo(rdb *redis.Client, ttl time.Duration) *LinkCache {
	return &LinkCache{rdb: rdb, linkTTL: ttl}
}

// Get retrieves the original URL for a short code from Redis cache.
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

// Set saves a shortCode-to-originalURL mapping into Redis cache with configured TTL.
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
