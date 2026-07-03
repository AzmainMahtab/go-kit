package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/elite4print/elite4print-go/internal/modules/identity/application"
	"github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	"github.com/elite4print/elite4print-go/internal/shared/eventbus"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
)

// UpdateUser handles user profile/status updates.
type UpdateUser struct {
	repo domain.UserRepository
	bus  eventbus.EventBus
	v    validator.Validator
}

// NewUpdateUser creates the use case.
func NewUpdateUser(repo domain.UserRepository, bus eventbus.EventBus, v validator.Validator) *UpdateUser {
	return &UpdateUser{repo: repo, bus: bus, v: v}
}

// Handle updates a user.
func (uc *UpdateUser) Handle(ctx context.Context, cmd application.UpdateUserCommand) (application.UserResult, error) {
	if err := uc.v.ValidateStruct(cmd); err != nil {
		return application.UserResult{}, err
	}

	user, err := uc.repo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return application.UserResult{}, fmt.Errorf("update user: %w", err)
	}
	if user == nil {
		return application.UserResult{}, domain.ErrUserNotFound
	}

	if cmd.FirstName != nil {
		user.FirstName = *cmd.FirstName
	}
	if cmd.LastName != nil {
		user.LastName = *cmd.LastName
	}
	if cmd.Phone != nil {
		user.Phone = *cmd.Phone
	}
	if cmd.CompanyName != nil {
		user.CompanyName = *cmd.CompanyName
	}
	if cmd.Status != nil {
		if err := user.TransitionStatus(*cmd.Status); err != nil {
			return application.UserResult{}, err
		}
	}

	user.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, user); err != nil {
		return application.UserResult{}, fmt.Errorf("update user: failed to persist: %w", err)
	}

	_ = uc.bus.Publish(ctx, domain.UserUpdatedEvent{
		UserID:     user.ID,
		Email:      user.Email.String(),
		Status:     string(user.Status),
		OccurredAt: time.Now().UTC(),
	})

	return application.ToUserResult(user), nil
}
