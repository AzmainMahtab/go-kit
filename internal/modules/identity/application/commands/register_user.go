package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/elite4print/elite4print-go/internal/modules/identity/application"
	"github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	"github.com/elite4print/elite4print-go/internal/shared/eventbus"
	"github.com/elite4print/elite4print-go/internal/shared/idgenerator"
	"github.com/elite4print/elite4print-go/internal/shared/password"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
)

// RegisterUser handles new user registration.
type RegisterUser struct {
	repo   domain.UserRepository
	bus    eventbus.EventBus
	hasher password.Hasher
	v      validator.Validator
}

// NewRegisterUser creates the use case.
func NewRegisterUser(repo domain.UserRepository, bus eventbus.EventBus, hasher password.Hasher, v validator.Validator) *RegisterUser {
	return &RegisterUser{repo: repo, bus: bus, hasher: hasher, v: v}
}

// Handle executes the registration command.
func (uc *RegisterUser) Handle(ctx context.Context, cmd application.RegisterUserCommand) (application.UserResult, error) {
	if err := uc.v.ValidateStruct(cmd); err != nil {
		return application.UserResult{}, err
	}

	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		return application.UserResult{}, err
	}

	exists, err := uc.repo.EmailExists(ctx, email)
	if err != nil {
		return application.UserResult{}, fmt.Errorf("register: %w", err)
	}
	if exists {
		return application.UserResult{}, domain.ErrUserAlreadyExists
	}

	hash, err := uc.hasher.Hash(cmd.Password)
	if err != nil {
		return application.UserResult{}, fmt.Errorf("register: failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           idgenerator.NewUUIDv7(),
		Email:        email,
		PasswordHash: password.NewHashed(hash),
		FirstName:    cmd.FirstName,
		LastName:     cmd.LastName,
		Phone:        cmd.Phone,
		CompanyName:  cmd.CompanyName,
		Role:         domain.Role(cmd.Role),
		Status:       domain.UserStatusPendingVerification,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.repo.Save(ctx, user); err != nil {
		return application.UserResult{}, fmt.Errorf("register: failed to save user: %w", err)
	}

	_ = uc.bus.Publish(ctx, domain.UserRegisteredEvent{
		UserID:     user.ID,
		Email:      user.Email.String(),
		Role:       string(user.Role),
		OccurredAt: now,
	})

	return application.ToUserResult(user), nil
}
