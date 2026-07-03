// Package application contains auth use cases and DTOs.
package application

import (
	"time"

	"github.com/google/uuid"
)

// -------------------- Commands --------------------

// LoginCommand is the input for login.
type LoginCommand struct {
	Email     string `validate:"required,email"`
	Password  string `validate:"required"`
	IP        string
	UserAgent string
}

// LogoutCommand is the input for logout.
type LogoutCommand struct {
	UserID    uuid.UUID `validate:"required"`
	SessionID uuid.UUID `validate:"required"`
	AccessJTI string    // optional: blacklist access token
}

// RefreshCommand is the input for token refresh.
type RefreshCommand struct {
	RefreshToken string `validate:"required"`
	IP           string
	UserAgent    string
}

// RevokeSessionCommand revokes a specific session.
type RevokeSessionCommand struct {
	UserID    uuid.UUID `validate:"required"`
	SessionID uuid.UUID `validate:"required"`
}

// RevokeAllSessionsCommand revokes all sessions except the current one.
type RevokeAllSessionsCommand struct {
	UserID          uuid.UUID `validate:"required"`
	CurrentSessionID uuid.UUID `validate:"required"`
}

// -------------------- Queries --------------------

// ListSessionsQuery lists active sessions for a user.
type ListSessionsQuery struct {
	UserID uuid.UUID `validate:"required"`
}

// GetSessionQuery fetches a single session.
type GetSessionQuery struct {
	SessionID uuid.UUID `validate:"required"`
}

// -------------------- Results --------------------

// TokenPairResult is returned on login/refresh.
type TokenPairResult struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	SessionID    uuid.UUID `json:"session_id"`
}

// SessionResult is the canonical session DTO.
type SessionResult struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	IP         string     `json:"ip"`
	UserAgent  string     `json:"user_agent"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	LastUsedAt time.Time  `json:"last_used_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	IsCurrent  bool       `json:"is_current"`
}
