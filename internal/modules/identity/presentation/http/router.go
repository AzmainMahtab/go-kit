package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewUserRouter mounts identity user routes.
// Registration is public; all other routes require authentication.
func NewUserRouter(handler *UserHandler, authMW func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Post("/register", handler.Register)

	r.Group(func(protected chi.Router) {
		if authMW != nil {
			protected.Use(authMW)
		}

		protected.Get("/", handler.List)
		protected.Get("/{id}", handler.GetByID)
		protected.Patch("/{id}", handler.Update)
	})

	return r
}
