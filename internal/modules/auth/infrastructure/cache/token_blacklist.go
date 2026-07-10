package cache

import (
	"context"
	"fmt"

	platformcache "github.com/example/go-kit/internal/platform/cache"
)

// RedisTokenBlacklist implements domain.TokenBlacklist.
type RedisTokenBlacklist struct {
	cache platformcache.Cache
}

// NewRedisTokenBlacklist creates a blacklist adapter.
func NewRedisTokenBlacklist(cache platformcache.Cache) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{cache: cache}
}

func blacklistKey(jti string) string {
	return fmt.Sprintf("blacklist:%s", jti)
}

// Add marks a token JTI as revoked.
func (r *RedisTokenBlacklist) Add(ctx context.Context, jti string, ttlSeconds int) error {
	return r.cache.Set(ctx, blacklistKey(jti), "1", ttlSeconds)
}

// IsBlacklisted reports whether a token JTI is revoked.
func (r *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	exists, err := r.cache.Exists(ctx, blacklistKey(jti))
	if err != nil {
		return false, err
	}
	return exists, nil
}
