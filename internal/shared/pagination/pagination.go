// Package pagination provides a reusable request/response model for list endpoints.
//
// Why not cursor-based pagination?
//   - Offset/limit is simpler for admin dashboards and matches common frontend
//     expectations.
//   - Cursor pagination can be added later for high-volume customer-facing lists.
package pagination

import (
	"math"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Request is the input for paginated queries.
type Request struct {
	Offset int `json:"offset" validate:"min=0"`
	Limit  int `json:"limit" validate:"min=1,max=100"`
}

// NewRequest returns a validated pagination request with safe defaults.
func NewRequest(offset, limit int) Request {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}
	return Request{Offset: offset, Limit: limit}
}

// Page is a generic paginated result.
type Page[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// TotalPages returns the total number of pages given the limit.
func (p Page[T]) TotalPages() int {
	if p.Limit == 0 {
		return 0
	}
	return int(math.Ceil(float64(p.Total) / float64(p.Limit)))
}

// IsEmpty reports whether the page has no items.
func (p Page[T]) IsEmpty() bool { return len(p.Items) == 0 }
