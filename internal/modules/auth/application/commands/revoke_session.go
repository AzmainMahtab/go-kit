package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/elite4print/elite4print-go/internal/modules/auth/application"
	"github.com/elite4print/elite4print-go/internal/modules/auth/domain"
	"github.com/elite4print/elite4print-go/internal/shared/eventbus"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
)

// RevokeSession revokes a single session owned by the user.
type RevokeSession struct {
	sessionRepo domain.SessionRepository
	cache       domain.SessionCache
	bus         eventbus.EventBus
	v           validator.Validator
}

// NewRevokeSession creates the use case.
func NewRevokeSession(
	sessionRepo domain.SessionRepository,
	cache domain.SessionCache,
	bus eventbus.EventBus,
	v validator.Validator,
) *RevokeSession {
	return &RevokeSession{sessionRepo: sessionRepo, cache: cache, bus: bus, v: v}
}

// Handle revokes the requested session.
func (uc *RevokeSession) Handle(ctx context.Context, cmd application.RevokeSessionCommand) error {
	if err := uc.v.ValidateStruct(cmd); err != nil {
		return err
	}

	session, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if session == nil || session.UserID != cmd.UserID {
		return domain.ErrSessionNotFound
	}

	if err := uc.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	_ = uc.cache.Delete(ctx, session.ID)

	_ = uc.bus.Publish(ctx, domain.UserLoggedOutEvent{
		UserID:     session.UserID,
		SessionID:  session.ID,
		Reason:     "revoked",
		OccurredAt: time.Now().UTC(),
	})

	return nil
}
