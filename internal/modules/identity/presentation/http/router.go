package http

import (
	"github.com/go-chi/chi/v5"
)

// NewUserRouter mounts identity user routes.
func NewUserRouter(handler *UserHandler) chi.Router {
	r := chi.NewRouter()

	r.Post("/register", handler.Register)
	r.Get("/", handler.List)
	r.Get("/{id}", handler.GetByID)
	r.Patch("/{id}", handler.Update)

	return r
}
