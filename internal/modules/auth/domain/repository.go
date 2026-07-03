package domain

import (
	"context"

	"github.com/google/uuid"
)

// SessionRepository is the port for session persistence.
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)
	GetByRefreshJTI(ctx context.Context, jti string) (*Session, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	Update(ctx context.Context, session *Session) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID, exceptSessionID uuid.UUID) error
}

// TokenBlacklist tracks revoked access/refresh tokens.
type TokenBlacklist interface {
	Add(ctx context.Context, jti string, ttlSeconds int) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}

// SessionCache provides fast session metadata lookups.
type SessionCache interface {
	Set(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, ttlSeconds int) error
	Get(ctx context.Context, sessionID uuid.UUID) (userID uuid.UUID, err error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
}
