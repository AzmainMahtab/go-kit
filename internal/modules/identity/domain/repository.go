package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository is the port used by the identity domain.
//
// Concrete implementations live in infrastructure/persistence. The domain
// depends only on this interface, which is the Dependency Inversion Principle.
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	EmailExists(ctx context.Context, email Email) (bool, error)
	Save(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	List(ctx context.Context, offset, limit int) ([]*User, int, error)
}
