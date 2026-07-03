// Package application contains the identity use cases and their DTOs.
//
// DTOs are plain structs. They travel from the HTTP layer into use cases.
package application

import (
	"time"

	"github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	"github.com/google/uuid"
)

// -------------------- Commands --------------------

// RegisterUserCommand is the input for user registration.
type RegisterUserCommand struct {
	Email           string `validate:"required,email"`
	Password        string `validate:"required,min=12"`
	ConfirmPassword string `validate:"required,eqfield=Password"`
	FirstName       string `validate:"required,max=100"`
	LastName        string `validate:"required,max=100"`
	Phone           string `validate:"omitempty,e164"`
	CompanyName     string `validate:"max=200"`
	Role            string `validate:"required,oneof=customer reseller admin management accountant customer_service operator hr"`
}

// UpdateUserCommand is the input for updating a user.
type UpdateUserCommand struct {
	UserID      uuid.UUID
	FirstName   *string
	LastName    *string
	Phone       *string
	CompanyName *string
	Status      *domain.UserStatus `validate:"omitempty,oneof=active pending_verification inactive suspended"`
}

// -------------------- Queries --------------------

// GetUserByIDQuery requests a single user.
type GetUserByIDQuery struct {
	UserID uuid.UUID `validate:"required"`
}

// ListUsersQuery requests a paginated list of users.
type ListUsersQuery struct {
	Offset int `validate:"min=0"`
	Limit  int `validate:"min=1,max=100"`
}

// -------------------- Results --------------------

// UserResult is the canonical user DTO returned by use cases.
type UserResult struct {
	ID            uuid.UUID         `json:"id"`
	Email         string            `json:"email"`
	FirstName     string            `json:"first_name"`
	LastName      string            `json:"last_name"`
	Phone         string            `json:"phone,omitempty"`
	CompanyName   string            `json:"company_name,omitempty"`
	Role          string            `json:"role"`
	Status        domain.UserStatus `json:"status"`
	EmailVerified bool              `json:"email_verified"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// ToUserResult maps a domain.User to a UserResult.
func ToUserResult(u *domain.User) UserResult {
	return UserResult{
		ID:            u.ID,
		Email:         u.Email.String(),
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Phone:         u.Phone,
		CompanyName:   u.CompanyName,
		Role:          string(u.Role),
		Status:        u.Status,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}
