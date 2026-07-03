package domain

import (
	"time"

	"github.com/elite4print/elite4print-go/internal/shared/eventbus"
	"github.com/google/uuid"
)

// UserLoggedInEvent is published on successful login.
type UserLoggedInEvent struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (UserLoggedInEvent) EventName() string { return "auth.user.logged_in" }

// UserLoggedOutEvent is published on logout/revoke.
type UserLoggedOutEvent struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	Reason    string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (UserLoggedOutEvent) EventName() string { return "auth.user.logged_out" }

var (
	_ eventbus.Event = UserLoggedInEvent{}
	_ eventbus.Event = UserLoggedOutEvent{}
)
