package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/elite4print/elite4print-go/internal/modules/auth/application"
	"github.com/elite4print/elite4print-go/internal/modules/auth/domain"
	identitydomain "github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	"github.com/elite4print/elite4print-go/internal/shared/eventbus"
	"github.com/elite4print/elite4print-go/internal/shared/idgenerator"
	"github.com/elite4print/elite4print-go/internal/shared/password"
	"github.com/elite4print/elite4print-go/internal/shared/token"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
)

// Login authenticates a user and creates a session.
type Login struct {
	userRepo    identitydomain.UserRepository
	sessionRepo domain.SessionRepository
	cache       domain.SessionCache
	blacklist   domain.TokenBlacklist
	tokenizer   token.Tokenizer
	hasher      password.Hasher
	bus         eventbus.EventBus
	v           validator.Validator
	cfg         LoginConfig
}

// LoginConfig holds TTLs for tokens.
type LoginConfig struct {
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	MaxConcurrent int
}

// NewLogin creates the login use case.
func NewLogin(
	userRepo identitydomain.UserRepository,
	sessionRepo domain.SessionRepository,
	cache domain.SessionCache,
	blacklist domain.TokenBlacklist,
	tokenizer token.Tokenizer,
	hasher password.Hasher,
	bus eventbus.EventBus,
	v validator.Validator,
	cfg LoginConfig,
) *Login {
	return &Login{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cache:       cache,
		blacklist:   blacklist,
		tokenizer:   tokenizer,
		hasher:      hasher,
		bus:         bus,
		v:           v,
		cfg:         cfg,
	}
}

// Handle executes login.
func (uc *Login) Handle(ctx context.Context, cmd application.LoginCommand) (application.TokenPairResult, error) {
	if err := uc.v.ValidateStruct(cmd); err != nil {
		return application.TokenPairResult{}, err
	}

	email, err := identitydomain.NewEmail(cmd.Email)
	if err != nil {
		return application.TokenPairResult{}, err
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return application.TokenPairResult{}, fmt.Errorf("login: %w", err)
	}
	if user == nil {
		return application.TokenPairResult{}, domain.ErrInvalidCredentials
	}

	if err := user.CanLogin(); err != nil {
		return application.TokenPairResult{}, err
	}

	if err := uc.hasher.Verify(cmd.Password, user.PasswordHash.String()); err != nil {
		return application.TokenPairResult{}, domain.ErrInvalidCredentials
	}

	// Enforce multi-session limit.
	sessions, err := uc.sessionRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return application.TokenPairResult{}, fmt.Errorf("login: %w", err)
	}
	activeCount := 0
	for _, s := range sessions {
		if s.IsActive(time.Now().UTC()) {
			activeCount++
		}
	}
	if uc.cfg.MaxConcurrent > 0 && activeCount >= uc.cfg.MaxConcurrent {
		// Revoke oldest active session to make room.
		if err := uc.revokeOldest(ctx, sessions); err != nil {
			return application.TokenPairResult{}, fmt.Errorf("login: %w", err)
		}
	}

	now := time.Now().UTC()
	session := &domain.Session{
		ID:         idgenerator.NewUUIDv7(),
		UserID:     user.ID,
		IP:         cmd.IP,
		UserAgent:  cmd.UserAgent,
		CreatedAt:  now,
		ExpiresAt:  now.Add(uc.cfg.RefreshTTL),
		LastUsedAt: now,
	}

	refreshToken, refreshJTI, err := uc.tokenizer.GenerateRefreshToken(session.ID, uc.cfg.RefreshTTL)
	if err != nil {
		return application.TokenPairResult{}, fmt.Errorf("login: %w", err)
	}
	session.RefreshJTI = refreshJTI

	accessToken, err := uc.tokenizer.GenerateAccessToken(user.ID, session.ID, user.Email.String(), string(user.Role), uc.cfg.AccessTTL)
	if err != nil {
		return application.TokenPairResult{}, fmt.Errorf("login: %w", err)
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return application.TokenPairResult{}, fmt.Errorf("login: %w", err)
	}

	_ = uc.cache.Set(ctx, session.ID, user.ID, int(uc.cfg.RefreshTTL.Seconds()))

	_ = uc.bus.Publish(ctx, domain.UserLoggedInEvent{
		UserID:     user.ID,
		SessionID:  session.ID,
		IP:         cmd.IP,
		UserAgent:  cmd.UserAgent,
		OccurredAt: now,
	})

	return application.TokenPairResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(uc.cfg.AccessTTL),
		SessionID:    session.ID,
	}, nil
}

func (uc *Login) revokeOldest(ctx context.Context, sessions []*domain.Session) error {
	var oldest *domain.Session
	for _, s := range sessions {
		if !s.IsActive(time.Now().UTC()) {
			continue
		}
		if oldest == nil || s.CreatedAt.Before(oldest.CreatedAt) {
			oldest = s
		}
	}
	if oldest == nil {
		return nil
	}
	return uc.sessionRepo.Revoke(ctx, oldest.ID)
}
