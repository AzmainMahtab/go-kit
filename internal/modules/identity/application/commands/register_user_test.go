package commands

import (
	"context"
	"testing"

	"github.com/example/go-kit/internal/modules/identity/application"
	"github.com/example/go-kit/internal/modules/identity/domain"
	"github.com/example/go-kit/internal/shared/eventbus"
	"github.com/example/go-kit/internal/shared/password"
	"github.com/example/go-kit/internal/shared/validator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// inMemoryUserRepository is a test fake.
type inMemoryUserRepository struct {
	users []*domain.User
}

func (r *inMemoryUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (r *inMemoryUserRepository) GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	for _, u := range r.users {
		if u.Email.String() == email.String() {
			return u, nil
		}
	}
	return nil, nil
}

func (r *inMemoryUserRepository) EmailExists(ctx context.Context, email domain.Email) (bool, error) {
	for _, u := range r.users {
		if u.Email.String() == email.String() {
			return true, nil
		}
	}
	return false, nil
}

func (r *inMemoryUserRepository) Save(ctx context.Context, user *domain.User) error {
	r.users = append(r.users, user)
	return nil
}

func (r *inMemoryUserRepository) Update(ctx context.Context, user *domain.User) error {
	for i, u := range r.users {
		if u.ID == user.ID {
			r.users[i] = user
			return nil
		}
	}
	return nil
}

func (r *inMemoryUserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, int, error) {
	return r.users, len(r.users), nil
}

func TestRegisterUser_Success(t *testing.T) {
	repo := &inMemoryUserRepository{}
	uc := NewRegisterUser(repo, eventbus.Noop{}, password.NewArgon2id(), validator.New())

	result, err := uc.Handle(context.Background(), application.RegisterUserCommand{
		Email:           "alice@example.com",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		FirstName:       "Alice",
		LastName:        "Smith",
		Role:            "customer",
	})

	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", result.Email)
	assert.Equal(t, "customer", result.Role)
	assert.Equal(t, domain.UserStatusPendingVerification, result.Status)
	assert.Len(t, repo.users, 1)
}

func TestRegisterUser_EmailAlreadyExists(t *testing.T) {
	repo := &inMemoryUserRepository{}
	uc := NewRegisterUser(repo, eventbus.Noop{}, password.NewArgon2id(), validator.New())

	cmd := application.RegisterUserCommand{
		Email:           "alice@example.com",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		FirstName:       "Alice",
		LastName:        "Smith",
		Role:            "customer",
	}

	_, err := uc.Handle(context.Background(), cmd)
	require.NoError(t, err)

	_, err = uc.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestRegisterUser_PasswordsDoNotMatch(t *testing.T) {
	repo := &inMemoryUserRepository{}
	uc := NewRegisterUser(repo, eventbus.Noop{}, password.NewArgon2id(), validator.New())

	_, err := uc.Handle(context.Background(), application.RegisterUserCommand{
		Email:           "alice@example.com",
		Password:        "StrongPass123!",
		ConfirmPassword: "DifferentPass123!",
		FirstName:       "Alice",
		LastName:        "Smith",
		Role:            "customer",
	})

	assert.Error(t, err)
}
