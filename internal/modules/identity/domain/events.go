package domain

import (
	"time"

	"github.com/example/go-kit/internal/shared/eventbus"
	"github.com/google/uuid"
)

// UserRegisteredEvent is published when a new user registers.
type UserRegisteredEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EventName returns the event name for routing.
func (UserRegisteredEvent) EventName() string { return "identity.user.registered" }

// UserUpdatedEvent is published when a user's core data changes.
type UserUpdatedEvent struct {
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	Status     string    `json:"status"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (UserUpdatedEvent) EventName() string { return "identity.user.updated" }

// Ensure events implement the Event marker interface.
var (
	_ eventbus.Event = UserRegisteredEvent{}
	_ eventbus.Event = UserUpdatedEvent{}
)
