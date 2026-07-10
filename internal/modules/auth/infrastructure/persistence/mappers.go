package persistence

import (
	"database/sql"
	"time"

	"github.com/example/go-kit/internal/modules/auth/domain"
	"github.com/google/uuid"
)

func sessionToDomain(m sessionModel) (*domain.Session, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(m.UserID)
	if err != nil {
		return nil, err
	}

	s := &domain.Session{
		ID:         id,
		UserID:     userID,
		RefreshJTI: m.RefreshJTI,
		IP:         m.IP,
		UserAgent:  m.UserAgent,
		CreatedAt:  m.CreatedAt,
		ExpiresAt:  m.ExpiresAt,
		LastUsedAt: m.LastUsedAt,
	}
	if m.RevokedAt.Valid {
		s.RevokedAt = &m.RevokedAt.Time
	}
	return s, nil
}

func sessionToModel(s *domain.Session) sessionModel {
	m := sessionModel{
		ID:         s.ID.String(),
		UserID:     s.UserID.String(),
		RefreshJTI: s.RefreshJTI,
		IP:         s.IP,
		UserAgent:  s.UserAgent,
		CreatedAt:  s.CreatedAt,
		ExpiresAt:  s.ExpiresAt,
		LastUsedAt: s.LastUsedAt,
	}
	if s.RevokedAt != nil {
		m.RevokedAt = sql.NullTime{Time: *s.RevokedAt, Valid: true}
	}
	return m
}

func now() time.Time { return time.Now().UTC() }
