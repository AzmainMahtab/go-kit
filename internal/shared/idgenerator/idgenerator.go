// Package idgenerator wraps UUIDv7 generation.
//
// UUIDv7 is preferred over UUIDv4 for database primary keys because it is
// time-ordered (roughly sequential), which improves B-tree index locality in
// PostgreSQL while still being globally unique and unguessable.
package idgenerator

import (
	"fmt"

	"github.com/google/uuid"
)

// NewUUIDv7 generates a new UUIDv7 or panics on failure.
//
// In practice uuid.NewV7 never fails, but we wrap it to keep the import count
// low and to make future changes easy.
func NewUUIDv7() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

// NewUUIDv7String returns the string form of a new UUIDv7.
func NewUUIDv7String() string {
	return NewUUIDv7().String()
}

// ParseUUID parses a UUID string.
func ParseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid: %w", err)
	}
	return id, nil
}
