// Package domain contains the identity bounded context's domain model.
//
// Domain code has zero dependencies on HTTP, SQL, Redis, or NATS.
package domain

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/elite4print/elite4print-go/internal/shared/password"
	"github.com/google/uuid"
)

// Role represents a user role in Elite4Print.
type Role string

const (
	RoleCustomer        Role = "customer"
	RoleReseller        Role = "reseller"
	RoleAdmin           Role = "admin"
	RoleManagement      Role = "management"
	RoleAccountant      Role = "accountant"
	RoleCustomerService Role = "customer_service"
	RoleOperator        Role = "operator"
	RoleHR              Role = "hr"
)

// UserStatus represents the lifecycle state of a user account.
type UserStatus string

const (
	UserStatusActive              UserStatus = "active"
	UserStatusPendingVerification UserStatus = "pending_verification"
	UserStatusInactive            UserStatus = "inactive"
	UserStatusSuspended           UserStatus = "suspended"
)

// User is the central identity aggregate root.
type User struct {
	ID            uuid.UUID
	Email         Email
	PasswordHash  password.Hashed // value object from shared/password
	FirstName     string
	LastName      string
	Phone         string
	CompanyName   string
	Role          Role
	Status        UserStatus
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// FullName returns the user's display name.
func (u *User) FullName() string {
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

// IsActive reports whether the user can authenticate.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && u.DeletedAt == nil
}

// CanLogin reports whether the user is allowed to log in.
func (u *User) CanLogin() error {
	if u.DeletedAt != nil {
		return ErrUserDeleted
	}
	switch u.Status {
	case UserStatusActive, UserStatusPendingVerification:
		return nil
	case UserStatusInactive:
		return ErrUserInactive
	case UserStatusSuspended:
		return ErrUserSuspended
	default:
		return ErrUserInactive
	}
}

// TransitionStatus moves the user to a new lifecycle state.
// It encodes the allowed transitions in the domain so handlers cannot set
// arbitrary statuses.
func (u *User) TransitionStatus(to UserStatus) error {
	allowed := map[UserStatus][]UserStatus{
		UserStatusPendingVerification: {UserStatusActive, UserStatusInactive, UserStatusSuspended},
		UserStatusActive:              {UserStatusInactive, UserStatusSuspended},
		UserStatusInactive:            {UserStatusActive},
		UserStatusSuspended:           {UserStatusActive},
	}

	for _, candidate := range allowed[u.Status] {
		if candidate == to {
			u.Status = to
			return nil
		}
	}

	return ErrInvalidStatusTransition
}

// Email is a value object.
type Email struct{ value string }

// NewEmail validates and creates an Email.
func NewEmail(v string) (Email, error) {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return Email{}, ErrEmailRequired
	}
	if _, err := mail.ParseAddress(v); err != nil {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: v}, nil
}

func (e Email) String() string { return e.value }

// Domain errors.
var (
	ErrEmailRequired           = errors.New("email is required")
	ErrInvalidEmail            = errors.New("invalid email address")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrUserInactive            = errors.New("user account is inactive")
	ErrUserSuspended           = errors.New("user account is suspended")
	ErrUserDeleted             = errors.New("user account is deleted")
	ErrWeakPassword            = errors.New("password does not meet strength requirements")
	ErrPasswordMismatch        = errors.New("passwords do not match")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)
