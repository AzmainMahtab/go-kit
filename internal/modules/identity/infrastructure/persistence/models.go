// Package persistence contains the SQL models and repository implementations
// for the identity module.
//
// These structs are NOT domain entities. They map 1:1 to the database schema.
package persistence

import (
	"database/sql"
	"time"
)

// userModel maps to the identity.users table.
type userModel struct {
	ID            string         `db:"id"`
	Email         string         `db:"email"`
	PasswordHash  string         `db:"password_hash"`
	FirstName     string         `db:"first_name"`
	LastName      string         `db:"last_name"`
	Phone         sql.NullString `db:"phone"`
	CompanyName   sql.NullString `db:"company_name"`
	Role          string         `db:"role"`
	Status        string         `db:"status"`
	EmailVerified bool           `db:"email_verified"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
	DeletedAt     sql.NullTime   `db:"deleted_at"`
}
