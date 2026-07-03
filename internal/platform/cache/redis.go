package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/elite4print/elite4print-go/internal/platform/config"
	"github.com/redis/go-redis/v9"
)

// Redis implements Cache using go-redis.
type Redis struct {
	client *redis.Client
}

// NewRedis creates a Redis cache from config.
func NewRedis(cfg *config.Config) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	return &Redis{client: client}
}

// Ping verifies connectivity.
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the connection.
func (r *Redis) Close() error {
	return r.client.Close()
}

// Client exposes the underlying go-redis client for health checks and
// advanced use cases.
func (r *Redis) Client() *redis.Client {
	return r.client
}

// Get retrieves a string value.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss{}
	}
	if err != nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

// Set stores a string value with a TTL in seconds.
func (r *Redis) Set(ctx context.Context, key string, value string, ttlSeconds int) error {
	return r.client.Set(ctx, key, value, time.Duration(ttlSeconds)*time.Second).Err()
}

// Delete removes a key.
func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists reports whether a key exists.
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}
