package http

import (
	"time"

	"github.com/example/go-kit/internal/modules/auth/application"
	"github.com/google/uuid"
)

// TokenPairResponse is the HTTP representation of a token pair.
type TokenPairResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	SessionID    string    `json:"session_id"`
}

// FromTokenPairResult maps an application result to an HTTP response.
func FromTokenPairResult(r application.TokenPairResult) TokenPairResponse {
	return TokenPairResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    r.ExpiresAt,
		SessionID:    r.SessionID.String(),
	}
}

// SessionResponse is the HTTP representation of a session.
type SessionResponse struct {
	ID         string    `json:"id"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	IsCurrent  bool      `json:"is_current"`
}

// FromSessionResult maps a session result.
func FromSessionResult(r application.SessionResult) SessionResponse {
	return SessionResponse{
		ID:         r.ID.String(),
		IP:         r.IP,
		UserAgent:  r.UserAgent,
		CreatedAt:  r.CreatedAt,
		ExpiresAt:  r.ExpiresAt,
		LastUsedAt: r.LastUsedAt,
		IsCurrent:  r.IsCurrent,
	}
}

// currentUserKey is used by auth middleware to store the current user in context.
type currentUserKey struct{}

// CurrentUser holds the authenticated caller's identity.
type CurrentUser struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	Email     string
	Role      string
}
