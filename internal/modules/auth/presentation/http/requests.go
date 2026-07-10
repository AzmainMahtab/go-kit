package http

import (
	"github.com/example/go-kit/internal/modules/auth/application"
)

// LoginRequest is the HTTP body for login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ToCommand maps the request to an application command.
func (r LoginRequest) ToCommand(ip, userAgent string) application.LoginCommand {
	return application.LoginCommand{
		Email:     r.Email,
		Password:  r.Password,
		IP:        ip,
		UserAgent: userAgent,
	}
}

// RefreshRequest is the HTTP body for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ToCommand maps the request to an application command.
func (r RefreshRequest) ToCommand(ip, userAgent string) application.RefreshCommand {
	return application.RefreshCommand{
		RefreshToken: r.RefreshToken,
		IP:           ip,
		UserAgent:    userAgent,
	}
}
