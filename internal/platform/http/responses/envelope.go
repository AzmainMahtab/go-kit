// Package responses provides the standard Elite4Print API response envelope.
//
// The existing Next.js and React dashboards expect this exact shape:
//
//	{
//	  "status_code": 200,
//	  "data": { ... },
//	  "count": 0,
//	  "total_pages": 0,
//	  "links": {},
//	  "message": "",
//	  "status": "success"
//	}
package responses

import (
	"encoding/json"
	"net/http"

	"github.com/elite4print/elite4print-go/internal/shared/apperrors"
	"github.com/elite4print/elite4print-go/internal/shared/pagination"
)

// Envelope is the standard API response wrapper.
type Envelope[T any] struct {
	StatusCode int    `json:"status_code"`
	Data       T      `json:"data,omitempty"`
	Count      int    `json:"count,omitempty"`
	TotalPages int    `json:"total_pages,omitempty"`
	Links      any    `json:"links,omitempty"`
	Message    string `json:"message,omitempty"`
	Status     string `json:"status"`
}

// Success creates a success envelope.
func Success[T any](statusCode int, data T) Envelope[T] {
	return Envelope[T]{
		StatusCode: statusCode,
		Data:       data,
		Status:     "success",
	}
}

// SuccessPage creates a success envelope for a paginated page.
func SuccessPage[T any](statusCode int, page pagination.Page[T]) Envelope[pagination.Page[T]] {
	return Envelope[pagination.Page[T]]{
		StatusCode: statusCode,
		Data:       page,
		Count:      page.Total,
		TotalPages: page.TotalPages(),
		Links:      map[string]string{},
		Status:     "success",
	}
}

// Error creates an error envelope.
func Error(statusCode int, code, message string) Envelope[any] {
	return Envelope[any]{
		StatusCode: statusCode,
		Message:    message,
		Status:     "error",
	}
}

// FromAppError converts an apperrors.AppError into an envelope.
func FromAppError(err *apperrors.AppError) Envelope[any] {
	return Error(err.StatusCode, err.Code, err.Message)
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// OK sends a 200 success response.
func OK[T any](w http.ResponseWriter, data T) {
	JSON(w, http.StatusOK, Success(http.StatusOK, data))
}

// Created sends a 201 success response.
func Created[T any](w http.ResponseWriter, data T) {
	JSON(w, http.StatusCreated, Success(http.StatusCreated, data))
}

// NoContent sends a 204 response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest sends a 400 validation error.
func BadRequest(w http.ResponseWriter, err error) {
	JSON(w, http.StatusBadRequest, Error(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
}

// Unauthorized sends a 401 error.
func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, Error(http.StatusUnauthorized, "UNAUTHORIZED", message))
}

// Forbidden sends a 403 error.
func Forbidden(w http.ResponseWriter, message string) {
	JSON(w, http.StatusForbidden, Error(http.StatusForbidden, "FORBIDDEN", message))
}

// NotFound sends a 404 error.
func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, Error(http.StatusNotFound, "NOT_FOUND", message))
}

// Conflict sends a 409 error.
func Conflict(w http.ResponseWriter, message string) {
	JSON(w, http.StatusConflict, Error(http.StatusConflict, "CONFLICT", message))
}

// InternalError sends a 500 error.
func InternalError(w http.ResponseWriter) {
	JSON(w, http.StatusInternalServerError, Error(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"))
}
