package queries

import (
	"context"
	"fmt"

	"github.com/elite4print/elite4print-go/internal/modules/identity/application"
	"github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
	"github.com/google/uuid"
)

// GetUser reads a single user.
type GetUser struct {
	repo domain.UserRepository
	v    validator.Validator
}

// NewGetUser creates the query handler.
func NewGetUser(repo domain.UserRepository, v validator.Validator) *GetUser {
	return &GetUser{repo: repo, v: v}
}

// ByID returns a user by ID.
func (q *GetUser) ByID(ctx context.Context, query application.GetUserByIDQuery) (application.UserResult, error) {
	if err := q.v.ValidateStruct(query); err != nil {
		return application.UserResult{}, err
	}

	user, err := q.repo.GetByID(ctx, query.UserID)
	if err != nil {
		return application.UserResult{}, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return application.UserResult{}, domain.ErrUserNotFound
	}
	return application.ToUserResult(user), nil
}

// ByIDRaw returns the raw domain user (used by other modules / ACL).
func (q *GetUser) ByIDRaw(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return q.repo.GetByID(ctx, id)
}
