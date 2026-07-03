// Package domain contains the auth bounded context's domain model.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session.
type Session struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	RefreshJTI   string // JWT ID of the refresh token
	IP           string
	UserAgent    string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastUsedAt   time.Time
	RevokedAt    *time.Time
}

// IsActive reports whether the session has not expired or been revoked.
func (s *Session) IsActive(now time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	return now.Before(s.ExpiresAt)
}

// Revoke marks the session as revoked.
func (s *Session) Revoke(now time.Time) {
	s.RevokedAt = &now
}

// Touch updates the last-used timestamp.
func (s *Session) Touch(now time.Time) {
	s.LastUsedAt = now
}

// Domain errors.
var (
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionExpired       = errors.New("session expired")
	ErrSessionRevoked       = errors.New("session revoked")
	ErrTokenBlacklisted     = errors.New("token has been revoked")
	ErrRefreshTokenMismatch = errors.New("refresh token mismatch")
	ErrUnauthorized         = errors.New("unauthorized")
)
