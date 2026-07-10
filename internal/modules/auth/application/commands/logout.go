package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/example/go-kit/internal/modules/auth/application"
	"github.com/example/go-kit/internal/modules/auth/domain"
	"github.com/example/go-kit/internal/shared/eventbus"
	"github.com/example/go-kit/internal/shared/validator"
)

// Logout revokes a session and blacklists the access token.
type Logout struct {
	sessionRepo domain.SessionRepository
	cache       domain.SessionCache
	blacklist   domain.TokenBlacklist
	bus         eventbus.EventBus
	v           validator.Validator
}

// NewLogout creates the logout use case.
func NewLogout(
	sessionRepo domain.SessionRepository,
	cache domain.SessionCache,
	blacklist domain.TokenBlacklist,
	bus eventbus.EventBus,
	v validator.Validator,
) *Logout {
	return &Logout{
		sessionRepo: sessionRepo,
		cache:       cache,
		blacklist:   blacklist,
		bus:         bus,
		v:           v,
	}
}

// Handle executes logout.
func (uc *Logout) Handle(ctx context.Context, cmd application.LogoutCommand) error {
	if err := uc.v.ValidateStruct(cmd); err != nil {
		return err
	}

	session, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	if session == nil || session.UserID != cmd.UserID {
		return domain.ErrSessionNotFound
	}

	if err := uc.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	_ = uc.cache.Delete(ctx, session.ID)

	if cmd.AccessJTI != "" {
		// Blacklist the access token for its remaining lifetime.
		_ = uc.blacklist.Add(ctx, cmd.AccessJTI, int(15*time.Minute.Seconds()))
	}

	_ = uc.bus.Publish(ctx, domain.UserLoggedOutEvent{
		UserID:     session.UserID,
		SessionID:  session.ID,
		Reason:     "logout",
		OccurredAt: time.Now().UTC(),
	})

	return nil
}
