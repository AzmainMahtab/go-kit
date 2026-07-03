package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/elite4print/elite4print-go/internal/modules/auth/application"
	"github.com/elite4print/elite4print-go/internal/modules/auth/domain"
	identitydomain "github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	"github.com/elite4print/elite4print-go/internal/shared/token"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
)

// Refresh issues a new access token given a valid refresh token.
type Refresh struct {
	userRepo    identitydomain.UserRepository
	sessionRepo domain.SessionRepository
	cache       domain.SessionCache
	blacklist   domain.TokenBlacklist
	tokenizer   token.Tokenizer
	v           validator.Validator
	cfg         RefreshConfig
}

// RefreshConfig holds token TTLs.
type RefreshConfig struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// NewRefresh creates the refresh use case.
func NewRefresh(
	userRepo identitydomain.UserRepository,
	sessionRepo domain.SessionRepository,
	cache domain.SessionCache,
	blacklist domain.TokenBlacklist,
	tokenizer token.Tokenizer,
	v validator.Validator,
	cfg RefreshConfig,
) *Refresh {
	return &Refresh{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cache:       cache,
		blacklist:   blacklist,
		tokenizer:   tokenizer,
		v:           v,
		cfg:         cfg,
	}
}

// Handle executes token refresh.
func (uc *Refresh) Handle(ctx context.Context, cmd application.RefreshCommand) (application.TokenPairResult, error) {
	if err := uc.v.ValidateStruct(cmd); err != nil {
		return application.TokenPairResult{}, err
	}

	sessionID, err := uc.tokenizer.ParseRefreshToken(ctx, cmd.RefreshToken)
	if err != nil {
		return application.TokenPairResult{}, domain.ErrUnauthorized
	}

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return application.TokenPairResult{}, fmt.Errorf("refresh: %w", err)
	}
	if session == nil {
		return application.TokenPairResult{}, domain.ErrSessionNotFound
	}

	now := time.Now().UTC()
	if !session.IsActive(now) {
		return application.TokenPairResult{}, domain.ErrSessionExpired
	}

	// Optional: verify refresh token JTI is not blacklisted.
	// For refresh tokens we typically just check session state.

	user, err := uc.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return application.TokenPairResult{}, fmt.Errorf("refresh: %w", err)
	}
	if user == nil {
		return application.TokenPairResult{}, domain.ErrUnauthorized
	}
	if err := user.CanLogin(); err != nil {
		return application.TokenPairResult{}, err
	}

	newRefreshToken, newRefreshJTI, err := uc.tokenizer.GenerateRefreshToken(session.ID, uc.cfg.RefreshTTL)
	if err != nil {
		return application.TokenPairResult{}, fmt.Errorf("refresh: %w", err)
	}

	session.Touch(now)
	session.RefreshJTI = newRefreshJTI
	if err := uc.sessionRepo.Update(ctx, session); err != nil {
		return application.TokenPairResult{}, fmt.Errorf("refresh: %w", err)
	}

	accessToken, err := uc.tokenizer.GenerateAccessToken(user.ID, session.ID, user.Email.String(), string(user.Role), uc.cfg.AccessTTL)
	if err != nil {
		return application.TokenPairResult{}, fmt.Errorf("refresh: %w", err)
	}

	_ = uc.cache.Set(ctx, session.ID, user.ID, int(uc.cfg.RefreshTTL.Seconds()))

	return application.TokenPairResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    now.Add(uc.cfg.AccessTTL),
		SessionID:    session.ID,
	}, nil
}
