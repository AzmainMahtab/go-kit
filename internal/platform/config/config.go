// Package config loads application configuration from environment variables.
// It uses github.com/caarlos0/env/v11 which is a tiny, reflection-based parser.
//
// Why env vars instead of a config file?
// - 12-Factor App compliance.
// - Docker and Kubernetes expect env-driven configuration.
// - No need to mount config files in production.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config is the single source of truth for all environment-driven settings.
type Config struct {
	// ---- Runtime -----------------------------------------------------------
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	HTTPPort    string `env:"HTTP_PORT" envDefault:"8080"`
	HTTPHost    string `env:"HTTP_HOST" envDefault:"0.0.0.0"`

	// ---- PostgreSQL --------------------------------------------------------
	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	PostgresPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"app"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"app"`
	PostgresDB       string `env:"POSTGRES_DB" envDefault:"appdb"`
	PostgresSSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`
	PostgresDSN      string `env:"POSTGRES_DSN" envDefault:""`

	// ---- Redis -------------------------------------------------------------
	RedisHost     string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort     int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`

	// ---- NATS --------------------------------------------------------------
	NATSURL string `env:"NATS_URL" envDefault:"nats://localhost:4222"`

	// ---- Security ----------------------------------------------------------
	JWTSecretKey         string        `env:"JWT_SECRET_KEY" envDefault:"dev-secret-replace-in-production"`
	JWTAccessTTL         time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	JWTRefreshTTL        time.Duration `env:"JWT_REFRESH_TTL" envDefault:"7d"`
	SessionMaxAge        time.Duration `env:"SESSION_MAX_AGE" envDefault:"7d"`
	SessionMaxConcurrent int           `env:"SESSION_MAX_CONCURRENT" envDefault:"5"`

	// ---- Rate limiting -----------------------------------------------------
	RateLimitRPS   int `env:"RATE_LIMIT_RPS" envDefault:"10"`
	RateLimitBurst int `env:"RATE_LIMIT_BURST" envDefault:"20"`
}

// Load parses environment variables into Config.
func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

// PostgresDSNOrBuild returns the explicit DSN if provided, otherwise builds one
// from the individual POSTGRES_* variables.
func (c *Config) PostgresDSNOrBuild() string {
	if c.PostgresDSN != "" {
		return c.PostgresDSN
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.PostgresUser, c.PostgresPassword, c.PostgresHost, c.PostgresPort, c.PostgresDB, c.PostgresSSLMode,
	)
}
