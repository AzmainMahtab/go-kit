package persistence

import (
	"database/sql"
	"time"

	"github.com/example/go-kit/internal/modules/identity/domain"
	"github.com/example/go-kit/internal/shared/password"
	"github.com/google/uuid"
)

func toDomain(m userModel) (*domain.User, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}

	email, err := domain.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:            id,
		Email:         email,
		PasswordHash:  password.NewHashed(m.PasswordHash),
		FirstName:     m.FirstName,
		LastName:      m.LastName,
		Phone:         m.Phone.String,
		CompanyName:   m.CompanyName.String,
		Role:          domain.Role(m.Role),
		Status:        domain.UserStatus(m.Status),
		EmailVerified: m.EmailVerified,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}

	if m.DeletedAt.Valid {
		user.DeletedAt = &m.DeletedAt.Time
	}

	return user, nil
}

func toModel(u *domain.User) userModel {
	m := userModel{
		ID:            u.ID.String(),
		Email:         u.Email.String(),
		PasswordHash:  u.PasswordHash.String(),
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Role:          string(u.Role),
		Status:        string(u.Status),
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}

	if u.Phone != "" {
		m.Phone = sql.NullString{String: u.Phone, Valid: true}
	}
	if u.CompanyName != "" {
		m.CompanyName = sql.NullString{String: u.CompanyName, Valid: true}
	}
	if u.DeletedAt != nil {
		m.DeletedAt = sql.NullTime{Time: *u.DeletedAt, Valid: true}
	}

	return m
}

func now() time.Time { return time.Now().UTC() }
