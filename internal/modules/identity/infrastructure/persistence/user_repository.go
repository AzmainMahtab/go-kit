package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/example/go-kit/internal/modules/identity/domain"
	"github.com/example/go-kit/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// PostgresUserRepository implements domain.UserRepository.
type PostgresUserRepository struct {
	tx *database.TxManager
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new repository.
func NewPostgresUserRepository(db *sqlx.DB, tx *database.TxManager) *PostgresUserRepository {
	return &PostgresUserRepository{tx: tx, db: db}
}

func (r *PostgresUserRepository) exec(ctx context.Context) database.Executor {
	if r.tx != nil {
		return r.tx.Executor(ctx)
	}
	return r.db
}

// GetByID returns a user by UUID.
func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var m userModel
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, company_name,
		       role, status, email_verified, created_at, updated_at, deleted_at
		FROM identity.users
		WHERE id = $1
	`
	if err := r.exec(ctx).GetContext(ctx, &m, query, id.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return toDomain(m)
}

// GetByEmail returns a user by email address.
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	var m userModel
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, company_name,
		       role, status, email_verified, created_at, updated_at, deleted_at
		FROM identity.users
		WHERE email = $1
	`
	if err := r.exec(ctx).GetContext(ctx, &m, query, email.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("GetByEmail: %w", err)
	}
	return toDomain(m)
}

// EmailExists checks whether an email is already registered.
func (r *PostgresUserRepository) EmailExists(ctx context.Context, email domain.Email) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM identity.users WHERE email = $1)`
	if err := r.exec(ctx).GetContext(ctx, &exists, query, email.String()); err != nil {
		return false, fmt.Errorf("EmailExists: %w", err)
	}
	return exists, nil
}

// Save inserts a new user.
func (r *PostgresUserRepository) Save(ctx context.Context, user *domain.User) error {
	m := toModel(user)
	query := `
		INSERT INTO identity.users
		(id, email, password_hash, first_name, last_name, phone, company_name,
		 role, status, email_verified, created_at, updated_at)
		VALUES
		(:id, :email, :password_hash, :first_name, :last_name, :phone, :company_name,
		 :role, :status, :email_verified, :created_at, :updated_at)
	`
	_, err := r.exec(ctx).NamedExecContext(ctx, query, m)
	if err != nil {
		return fmt.Errorf("Save: %w", err)
	}
	return nil
}

// Update updates an existing user.
func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	m := toModel(user)
	m.UpdatedAt = now()
	query := `
		UPDATE identity.users
		SET first_name = :first_name,
		    last_name = :last_name,
		    phone = :phone,
		    company_name = :company_name,
		    role = :role,
		    status = :status,
		    email_verified = :email_verified,
		    updated_at = :updated_at
		WHERE id = :id
	`
	_, err := r.exec(ctx).NamedExecContext(ctx, query, m)
	if err != nil {
		return fmt.Errorf("Update: %w", err)
	}
	return nil
}

// List returns a paginated list of non-deleted users.
func (r *PostgresUserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM identity.users WHERE deleted_at IS NULL`
	if err := r.exec(ctx).GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("List count: %w", err)
	}

	var models []userModel
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, company_name,
		       role, status, email_verified, created_at, updated_at, deleted_at
		FROM identity.users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		OFFSET $1 LIMIT $2
	`
	if err := r.exec(ctx).SelectContext(ctx, &models, query, offset, limit); err != nil {
		return nil, 0, fmt.Errorf("List: %w", err)
	}

	users := make([]*domain.User, len(models))
	for i, m := range models {
		u, err := toDomain(m)
		if err != nil {
			return nil, 0, fmt.Errorf("List: mapping failed: %w", err)
		}
		users[i] = u
	}
	return users, total, nil
}
