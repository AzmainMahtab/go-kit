package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/elite4print/elite4print-go/internal/modules/auth/domain"
	"github.com/elite4print/elite4print-go/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// PostgresSessionRepository implements domain.SessionRepository.
type PostgresSessionRepository struct {
	tx *database.TxManager
	db *sqlx.DB
}

// NewPostgresSessionRepository creates a repository.
func NewPostgresSessionRepository(db *sqlx.DB, tx *database.TxManager) *PostgresSessionRepository {
	return &PostgresSessionRepository{tx: tx, db: db}
}

func (r *PostgresSessionRepository) exec(ctx context.Context) database.Executor {
	if r.tx != nil {
		return r.tx.Executor(ctx)
	}
	return r.db
}

// Create inserts a session.
func (r *PostgresSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	m := sessionToModel(session)
	query := `
		INSERT INTO identity.sessions
		(id, user_id, refresh_jti, ip, user_agent, created_at, expires_at, last_used_at)
		VALUES
		(:id, :user_id, :refresh_jti, :ip, :user_agent, :created_at, :expires_at, :last_used_at)
	`
	_, err := r.exec(ctx).NamedExecContext(ctx, query, m)
	if err != nil {
		return fmt.Errorf("Create session: %w", err)
	}
	return nil
}

// GetByID fetches a session by ID.
func (r *PostgresSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	var m sessionModel
	query := `
		SELECT id, user_id, refresh_jti, ip, user_agent, created_at, expires_at, last_used_at, revoked_at
		FROM identity.sessions
		WHERE id = $1
	`
	if err := r.exec(ctx).GetContext(ctx, &m, query, id.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("GetByID session: %w", err)
	}
	return sessionToDomain(m)
}

// GetByRefreshJTI fetches a session by refresh token JTI.
func (r *PostgresSessionRepository) GetByRefreshJTI(ctx context.Context, jti string) (*domain.Session, error) {
	var m sessionModel
	query := `
		SELECT id, user_id, refresh_jti, ip, user_agent, created_at, expires_at, last_used_at, revoked_at
		FROM identity.sessions
		WHERE refresh_jti = $1
	`
	if err := r.exec(ctx).GetContext(ctx, &m, query, jti); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("GetByRefreshJTI session: %w", err)
	}
	return sessionToDomain(m)
}

// ListByUserID lists all sessions for a user.
func (r *PostgresSessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	var models []sessionModel
	query := `
		SELECT id, user_id, refresh_jti, ip, user_agent, created_at, expires_at, last_used_at, revoked_at
		FROM identity.sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	if err := r.exec(ctx).SelectContext(ctx, &models, query, userID.String()); err != nil {
		return nil, fmt.Errorf("ListByUserID sessions: %w", err)
	}

	sessions := make([]*domain.Session, len(models))
	for i, m := range models {
		s, err := sessionToDomain(m)
		if err != nil {
			return nil, fmt.Errorf("ListByUserID sessions: mapping failed: %w", err)
		}
		sessions[i] = s
	}
	return sessions, nil
}

// Update updates a session.
func (r *PostgresSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	m := sessionToModel(session)
	m.LastUsedAt = now()
	query := `
		UPDATE identity.sessions
		SET last_used_at = :last_used_at,
		    refresh_jti = :refresh_jti
		WHERE id = :id
	`
	_, err := r.exec(ctx).NamedExecContext(ctx, query, m)
	if err != nil {
		return fmt.Errorf("Update session: %w", err)
	}
	return nil
}

// Revoke marks a session as revoked.
func (r *PostgresSessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE identity.sessions SET revoked_at = $1 WHERE id = $2`
	_, err := r.exec(ctx).ExecContext(ctx, query, now(), id.String())
	if err != nil {
		return fmt.Errorf("Revoke session: %w", err)
	}
	return nil
}

// RevokeAllForUser revokes all sessions for a user except one.
func (r *PostgresSessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID, exceptSessionID uuid.UUID) error {
	query := `
		UPDATE identity.sessions
		SET revoked_at = $1
		WHERE user_id = $2 AND id != $3 AND revoked_at IS NULL
	`
	_, err := r.exec(ctx).ExecContext(ctx, query, now(), userID.String(), exceptSessionID.String())
	if err != nil {
		return fmt.Errorf("RevokeAllForUser sessions: %w", err)
	}
	return nil
}
