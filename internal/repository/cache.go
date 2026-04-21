package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Ozdal97/go-url-shortener/internal/domain"
)

type LinkCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewLinkCache(rdb *redis.Client, ttl time.Duration) *LinkCache {
	return &LinkCache{rdb: rdb, ttl: ttl}
}

func (c *LinkCache) Get(ctx context.Context, code string) (string, error) {
	v, err := c.rdb.Get(ctx, key(code)).Result()
	if errors.Is(err, redis.Nil) {
		return "", domain.ErrNotFound
	}
	return v, err
}

func (c *LinkCache) Set(ctx context.Context, code, target string) error {
	return c.rdb.Set(ctx, key(code), target, c.ttl).Err()
}

func (c *LinkCache) Del(ctx context.Context, code string) error {
	return c.rdb.Del(ctx, key(code)).Err()
}

func key(code string) string { return "sl:" + code }
