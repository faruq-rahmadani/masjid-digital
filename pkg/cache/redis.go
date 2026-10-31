package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis stores keys in a Redis-compatible server (Redis, Valkey, Dragonfly, KeyDB)
type Redis struct {
	client *redis.Client
}

var _ Cache = (*Redis)(nil)

// NewRedis connects to the server at dsn (redis://[:password@]host:port[/db])
// and verifies it is reachable
func NewRedis(ctx context.Context, dsn string) (*Redis, error) {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing redis dsn: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("error connecting to redis: %w", err)
	}

	return &Redis{client: client}, nil
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("error getting cache key: %w", err)
	}

	return value, nil
}

func (r *Redis) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	// go-redis reads a negative ttl as KEEPTTL; treat it as no expiry like the memory driver
	if ttl < 0 {
		ttl = 0
	}

	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("error setting cache key: %w", err)
	}

	return nil
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	// DEL without keys is a Redis syntax error
	if len(keys) == 0 {
		return nil
	}

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("error deleting cache keys: %w", err)
	}

	return nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}
