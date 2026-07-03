// Package health reports the status of external dependencies.
package health

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// Status describes the state of a single dependency.
type Status string

const (
	StatusUp   Status = "up"
	StatusDown Status = "down"
)

// Result is the aggregated health result.
type Result struct {
	Status     Status            `json:"status"`
	Components map[string]Status `json:"components"`
}

// Checker verifies the availability of backing services.
type Checker struct {
	db    *sqlx.DB
	redis *redis.Client
}

// NewChecker creates a health checker.
func NewChecker(db *sqlx.DB, redis *redis.Client) *Checker {
	return &Checker{db: db, redis: redis}
}

// Check returns the current health of all dependencies.
func (c *Checker) Check(ctx context.Context) Result {
	components := map[string]Status{}
	overall := StatusUp

	if c.db != nil {
		if err := c.db.PingContext(ctx); err != nil {
			components["postgres"] = StatusDown
			overall = StatusDown
		} else {
			components["postgres"] = StatusUp
		}
	} else {
		components["postgres"] = StatusDown
		overall = StatusDown
	}

	if c.redis != nil {
		if err := c.redis.Ping(ctx).Err(); err != nil {
			components["redis"] = StatusDown
			overall = StatusDown
		} else {
			components["redis"] = StatusUp
		}
	} else {
		components["redis"] = StatusDown
		overall = StatusDown
	}

	return Result{Status: overall, Components: components}
}
