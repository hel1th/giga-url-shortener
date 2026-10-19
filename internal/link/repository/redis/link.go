// Package redis provides Redis caching for resolved link URLs.
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const (
	tracerName = "repo.redis"

	linkKeyPrefix = "link:"
)

var tracer = otel.Tracer(tracerName)

// LinkCache implements usecase.LinkCache backed by Redis.
type LinkCache struct {
	rdb              *redis.Client
	linkTTL          time.Duration
	negativeCacheTTL time.Duration
}

// NewRedisRepo creates a new LinkCache with the provided Redis client and TTL duration.
func NewRedisRepo(rdb *redis.Client, ttl, negativeCacheTTL time.Duration) *LinkCache {
	return &LinkCache{rdb: rdb, linkTTL: ttl, negativeCacheTTL: negativeCacheTTL}
}

// Get retrieves the original URL for a short code from Redis cache.
func (c *LinkCache) Get(ctx context.Context, shortCode string) (string, error) {
	ctx, span := tracer.Start(ctx, "Get")
	defer span.End()

	key := getLinkKey(shortCode)

	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		if val == "" {
			return "", domain.ErrNotFoundCached
		}
		return val, nil
	}

	if errors.Is(err, redis.Nil) {
		return "", domain.ErrNotFound
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return "", fmt.Errorf("failed to get link %s from redis: %w", shortCode, err)
}

// Set saves a shortCode-to-originalURL mapping into Redis cache with configured TTL.
func (c *LinkCache) Set(ctx context.Context, shortCode, originalURL string) error {
	ctx, span := tracer.Start(ctx, "Set")
	defer span.End()

	key := getLinkKey(shortCode)

	err := c.rdb.Set(ctx, key, originalURL, c.linkTTL).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to set link %s to redis: %w", shortCode, err)
	}

	return nil
}

// SetUnavailable saves a shortCode to "" mapping into Redis cache with configured negativeCacheTTL.
func (c *LinkCache) SetUnavailable(ctx context.Context, shortCode string) error {
	ctx, span := tracer.Start(ctx, "SetUnavailable")
	defer span.End()

	key := getLinkKey(shortCode)

	err := c.rdb.Set(ctx, key, "", c.negativeCacheTTL).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to set link %s to redis: %w", shortCode, err)
	}

	return nil
}

func getLinkKey(shortCode string) string {
	return linkKeyPrefix + shortCode
}
