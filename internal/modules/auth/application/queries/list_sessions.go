package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/elite4print/elite4print-go/internal/modules/auth/application"
	"github.com/elite4print/elite4print-go/internal/modules/auth/domain"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
	"github.com/google/uuid"
)

// ListSessions lists active sessions for a user.
type ListSessions struct {
	repo domain.SessionRepository
	v    validator.Validator
}

// NewListSessions creates the query handler.
func NewListSessions(repo domain.SessionRepository, v validator.Validator) *ListSessions {
	return &ListSessions{repo: repo, v: v}
}

// Execute runs the query.
func (q *ListSessions) Execute(ctx context.Context, query application.ListSessionsQuery, currentSessionID uuid.UUID) ([]application.SessionResult, error) {
	if err := q.v.ValidateStruct(query); err != nil {
		return nil, err
	}

	sessions, err := q.repo.ListByUserID(ctx, query.UserID)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	now := time.Now().UTC()
	results := make([]application.SessionResult, 0, len(sessions))
	for _, s := range sessions {
		if !s.IsActive(now) {
			continue
		}
		results = append(results, application.SessionResult{
			ID:         s.ID,
			UserID:     s.UserID,
			IP:         s.IP,
			UserAgent:  s.UserAgent,
			CreatedAt:  s.CreatedAt,
			ExpiresAt:  s.ExpiresAt,
			LastUsedAt: s.LastUsedAt,
			RevokedAt:  s.RevokedAt,
			IsCurrent:  s.ID == currentSessionID,
		})
	}

	return results, nil
}
