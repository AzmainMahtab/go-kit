package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/example/go-kit/internal/modules/auth/domain"
	"github.com/example/go-kit/internal/platform/http/responses"
	"github.com/example/go-kit/internal/shared/token"
	"github.com/google/uuid"
)

// accessJTIKey is used to store the access token JTI in request context.
type accessJTIKey struct{}

// AuthMiddleware validates access tokens and loads the current user.
type AuthMiddleware struct {
	tokenizer token.Tokenizer
	blacklist domain.TokenBlacklist
}

// NewAuthMiddleware creates the middleware.
func NewAuthMiddleware(tokenizer token.Tokenizer, blacklist domain.TokenBlacklist) *AuthMiddleware {
	return &AuthMiddleware{tokenizer: tokenizer, blacklist: blacklist}
}

// Authenticate returns a middleware that requires a valid access token.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := extractBearerToken(r)
		if !ok {
			responses.Unauthorized(w, "missing or invalid authorization header")
			return
		}

		claims, err := m.tokenizer.ParseAccessToken(r.Context(), tokenString)
		if err != nil {
			responses.Unauthorized(w, "invalid or expired token")
			return
		}

		// Check blacklist.
		if claims.ID != "" {
			blacklisted, err := m.blacklist.IsBlacklisted(r.Context(), claims.ID)
			if err != nil {
				responses.InternalError(w)
				return
			}
			if blacklisted {
				responses.Unauthorized(w, "token revoked")
				return
			}
		}

		ctx := context.WithValue(r.Context(), currentUserKey{}, CurrentUser{
			UserID:    claims.UserID,
			SessionID: claims.SessionID,
			Email:     claims.Email,
			Role:      claims.Role,
		})
		ctx = context.WithValue(ctx, accessJTIKey{}, claims.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CurrentUserFromContext extracts the current user from context.
func CurrentUserFromContext(ctx context.Context) (CurrentUser, bool) {
	if user, ok := ctx.Value(currentUserKey{}).(CurrentUser); ok {
		return user, true
	}
	return CurrentUser{UserID: uuid.Nil, SessionID: uuid.Nil}, false
}

func extractBearerToken(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if auth == "" || !strings.HasPrefix(auth, prefix) {
		return "", false
	}
	return strings.TrimPrefix(auth, prefix), true
}
