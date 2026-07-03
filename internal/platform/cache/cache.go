// Package cache defines the cache port used by auth and other modules.
//
// Redis is used for:
// - Session metadata and token blacklists.
// - Short-lived caches (user profile, OTP).
// - Rate-limit counters.
package cache

import "context"

// Cache is a minimal key-value cache interface.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttlSeconds int) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// ErrCacheMiss is returned when a key is not found.
type ErrCacheMiss struct{}

func (ErrCacheMiss) Error() string { return "cache miss" }
