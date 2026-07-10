package middleware

import (
	"net/http"

	"github.com/example/go-kit/internal/platform/http/responses"
	"github.com/example/go-kit/internal/shared/logger"
)

// Recovery recovers from panics and returns a 500 response.
func Recovery(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						"error", rec,
						"path", r.URL.Path,
						"request_id", GetRequestID(r.Context()),
					)
					responses.InternalError(w)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
