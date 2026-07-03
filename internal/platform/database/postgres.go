// Package database provides PostgreSQL connectivity using sqlx over pgx.
//
// Why sqlx + pgx?
// - pgx is the fastest and most feature-complete Postgres driver for Go.
// - sqlx adds convenient struct scanning and named queries on top.
// - Together they give full SQL control without the magic of an ORM.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/elite4print/elite4print-go/internal/platform/config"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
	"github.com/jmoiron/sqlx"
)

// NewPool returns a configured *sqlx.DB.
func NewPool(cfg *config.Config) (*sqlx.DB, error) {
	dsn := cfg.PostgresDSNOrBuild()

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return db, nil
}
