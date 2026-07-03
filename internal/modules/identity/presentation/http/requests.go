package http

import (
	"github.com/elite4print/elite4print-go/internal/modules/identity/application"
	"github.com/elite4print/elite4print-go/internal/modules/identity/domain"
)

// RegisterUserRequest is the HTTP body for registration.
type RegisterUserRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=12"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	FirstName       string `json:"first_name" validate:"required,max=100"`
	LastName        string `json:"last_name" validate:"required,max=100"`
	Phone           string `json:"phone,omitempty" validate:"omitempty,e164"`
	CompanyName     string `json:"company_name,omitempty" validate:"max=200"`
	Role            string `json:"role" validate:"required,oneof=customer reseller admin management accountant customer_service operator hr"`
}

// ToCommand maps the HTTP request to the application command.
func (r RegisterUserRequest) ToCommand() application.RegisterUserCommand {
	return application.RegisterUserCommand{
		Email:           r.Email,
		Password:        r.Password,
		ConfirmPassword: r.ConfirmPassword,
		FirstName:       r.FirstName,
		LastName:        r.LastName,
		Phone:           r.Phone,
		CompanyName:     r.CompanyName,
		Role:            r.Role,
	}
}

// UpdateUserRequest is the HTTP body for updates.
type UpdateUserRequest struct {
	FirstName   *string `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName    *string `json:"last_name,omitempty" validate:"omitempty,max=100"`
	Phone       *string `json:"phone,omitempty" validate:"omitempty,e164"`
	CompanyName *string `json:"company_name,omitempty" validate:"omitempty,max=200"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active pending_verification inactive suspended"`
}

// ToCommand maps the request to an application command.
func (r UpdateUserRequest) ToCommand() application.UpdateUserCommand {
	cmd := application.UpdateUserCommand{
		FirstName:   r.FirstName,
		LastName:    r.LastName,
		Phone:       r.Phone,
		CompanyName: r.CompanyName,
	}
	if r.Status != nil {
		status := domain.UserStatus(*r.Status)
		cmd.Status = &status
	}
	return cmd
}
