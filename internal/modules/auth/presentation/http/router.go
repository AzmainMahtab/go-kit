package http

import (
	"github.com/go-chi/chi/v5"
)

// NewAuthRouter mounts auth routes.
// Public routes (login/refresh) are mounted directly. Protected routes use the
// auth middleware.
func NewAuthRouter(handler *AuthHandler, authMW *AuthMiddleware) chi.Router {
	r := chi.NewRouter()

	r.Post("/login", handler.Login)
	r.Post("/refresh", handler.Refresh)

	r.Group(func(protected chi.Router) {
		protected.Use(authMW.Authenticate)

		protected.Post("/logout", handler.Logout)
		protected.Get("/sessions", handler.ListSessions)
		protected.Post("/sessions/revoke-all", handler.RevokeAllSessions)
		protected.Post("/sessions/{id}/revoke", handler.RevokeSession)
	})

	return r
}
