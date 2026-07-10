package commands

import (
	"context"
	"time"

	"github.com/example/go-kit/internal/modules/auth/application"
	"github.com/example/go-kit/internal/modules/auth/domain"
	"github.com/example/go-kit/internal/shared/eventbus"
	"github.com/example/go-kit/internal/shared/validator"
)

// RevokeAllSessions revokes every session except the current one.
type RevokeAllSessions struct {
	sessionRepo domain.SessionRepository
	cache       domain.SessionCache
	bus         eventbus.EventBus
	v           validator.Validator
}

// NewRevokeAllSessions creates the use case.
func NewRevokeAllSessions(
	sessionRepo domain.SessionRepository,
	cache domain.SessionCache,
	bus eventbus.EventBus,
	v validator.Validator,
) *RevokeAllSessions {
	return &RevokeAllSessions{sessionRepo: sessionRepo, cache: cache, bus: bus, v: v}
}

// Handle revokes all sessions except current.
func (uc *RevokeAllSessions) Handle(ctx context.Context, cmd application.RevokeAllSessionsCommand) error {
	if err := uc.v.ValidateStruct(cmd); err != nil {
		return err
	}

	if err := uc.sessionRepo.RevokeAllForUser(ctx, cmd.UserID, cmd.CurrentSessionID); err != nil {
		return err
	}

	// Clear cache entries for all revoked sessions (best-effort).
	sessions, err := uc.sessionRepo.ListByUserID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	for _, s := range sessions {
		if s.ID != cmd.CurrentSessionID {
			_ = uc.cache.Delete(ctx, s.ID)
		}
	}

	_ = uc.bus.Publish(ctx, domain.UserLoggedOutEvent{
		UserID:     cmd.UserID,
		SessionID:  cmd.CurrentSessionID,
		Reason:     "revoke_all_others",
		OccurredAt: time.Now().UTC(),
	})

	return nil
}
