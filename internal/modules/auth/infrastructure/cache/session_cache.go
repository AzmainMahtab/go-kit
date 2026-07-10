// Package cache contains Redis-backed adapters for auth domain ports.
package cache

import (
	"context"
	"fmt"

	"github.com/example/go-kit/internal/modules/auth/domain"
	platformcache "github.com/example/go-kit/internal/platform/cache"
	"github.com/google/uuid"
)

// RedisSessionCache implements domain.SessionCache.
type RedisSessionCache struct {
	cache platformcache.Cache
}

// NewRedisSessionCache creates a session cache adapter.
func NewRedisSessionCache(cache platformcache.Cache) *RedisSessionCache {
	return &RedisSessionCache{cache: cache}
}

func key(sessionID uuid.UUID) string {
	return fmt.Sprintf("session:%s", sessionID.String())
}

// Set stores session metadata in Redis.
func (r *RedisSessionCache) Set(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, ttlSeconds int) error {
	return r.cache.Set(ctx, key(sessionID), userID.String(), ttlSeconds)
}

// Get returns the user ID associated with a session.
func (r *RedisSessionCache) Get(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	val, err := r.cache.Get(ctx, key(sessionID))
	if err != nil {
		if _, ok := err.(platformcache.ErrCacheMiss); ok {
			return uuid.Nil, domain.ErrSessionNotFound
		}
		return uuid.Nil, err
	}
	return uuid.Parse(val)
}

// Delete removes a session from cache.
func (r *RedisSessionCache) Delete(ctx context.Context, sessionID uuid.UUID) error {
	return r.cache.Delete(ctx, key(sessionID))
}
