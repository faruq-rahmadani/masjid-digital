// Package cache is a key-value store with expiry behind one interface, backed
// either by process memory or by Redis, so callers never need to know which.
//
// The memory driver needs nothing extra but loses every key on restart and is
// not shared between app instances. Use Redis (or Valkey, Dragonfly, KeyDB)
// when running more than one instance or when keys must survive restarts.
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Driver is a supported cache backend
type Driver string

const (
	DriverMemory Driver = "memory"
	DriverRedis  Driver = "redis"
)

// ErrNotFound is returned by Get for a key that is missing or expired
var ErrNotFound = errors.New("cache: key not found")

// janitorInterval is how often the memory driver frees expired keys
const janitorInterval = time.Minute

type Cache interface {
	// Get returns the value of key, or ErrNotFound when it is missing or expired
	Get(ctx context.Context, key string) (string, error)
	// Set stores value under key. A ttl of zero or less means it never expires.
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	// Delete removes keys; missing keys are ignored
	Delete(ctx context.Context, keys ...string) error
	// Close releases the cache's resources
	Close() error
}

// Config selects the cache backend
type Config struct {
	Driver Driver
	DSN    string // redis only, e.g. redis://:password@localhost:6379/0
}

// Open returns the cache described by cfg, verifying Redis is reachable
func Open(ctx context.Context, cfg Config) (Cache, error) {
	switch cfg.Driver {
	case DriverMemory:
		return NewMemory(janitorInterval), nil
	case DriverRedis:
		return NewRedis(ctx, cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported cache driver %q, must be one of: %s, %s", cfg.Driver, DriverMemory, DriverRedis)
	}
}
