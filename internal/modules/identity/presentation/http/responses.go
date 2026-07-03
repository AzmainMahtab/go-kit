package http

import (
	"github.com/elite4print/elite4print-go/internal/modules/identity/application"
)

// UserResponse is the HTTP representation of a user.
type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Phone         string `json:"phone,omitempty"`
	CompanyName   string `json:"company_name,omitempty"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// FromUserResult maps an application result to an HTTP response.
func FromUserResult(r application.UserResult) UserResponse {
	return UserResponse{
		ID:            r.ID.String(),
		Email:         r.Email,
		FirstName:     r.FirstName,
		LastName:      r.LastName,
		Phone:         r.Phone,
		CompanyName:   r.CompanyName,
		Role:          r.Role,
		Status:        string(r.Status),
		EmailVerified: r.EmailVerified,
		CreatedAt:     r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ListUsersResponse wraps a paginated list.
type ListUsersResponse struct {
	Items      []UserResponse `json:"items"`
	Total      int            `json:"total"`
	Offset     int            `json:"offset"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}
