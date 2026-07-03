// Package persistence contains the auth module's SQL models and repositories.
package persistence

import (
	"database/sql"
	"time"
)

// sessionModel maps to the identity.sessions table.
type sessionModel struct {
	ID         string       `db:"id"`
	UserID     string       `db:"user_id"`
	RefreshJTI string       `db:"refresh_jti"`
	IP         string       `db:"ip"`
	UserAgent  string       `db:"user_agent"`
	CreatedAt  time.Time    `db:"created_at"`
	ExpiresAt  time.Time    `db:"expires_at"`
	LastUsedAt time.Time    `db:"last_used_at"`
	RevokedAt  sql.NullTime `db:"revoked_at"`
}
