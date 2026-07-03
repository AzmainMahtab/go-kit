// Package auth wires the authentication bounded context.
package auth

import (
	"net/http"

	"github.com/elite4print/elite4print-go/internal/modules/auth/application/commands"
	"github.com/elite4print/elite4print-go/internal/modules/auth/application/queries"
	"github.com/elite4print/elite4print-go/internal/modules/auth/domain"
	authcache "github.com/elite4print/elite4print-go/internal/modules/auth/infrastructure/cache"
	"github.com/elite4print/elite4print-go/internal/modules/auth/infrastructure/persistence"
	authHTTP "github.com/elite4print/elite4print-go/internal/modules/auth/presentation/http"
	identitydomain "github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	platformcache "github.com/elite4print/elite4print-go/internal/platform/cache"
	"github.com/elite4print/elite4print-go/internal/platform/config"
	"github.com/elite4print/elite4print-go/internal/platform/database"
	"github.com/elite4print/elite4print-go/internal/shared/eventbus"
	"github.com/elite4print/elite4print-go/internal/shared/password"
	"github.com/elite4print/elite4print-go/internal/shared/token"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// Module is the public API of the auth bounded context.
type Module struct {
	authRouter chi.Router
	authMW     *authHTTP.AuthMiddleware
}

// Deps contains the external dependencies auth needs.
type Deps struct {
	DB             *sqlx.DB
	Tx             *database.TxManager
	Cache          platformcache.Cache
	Bus            eventbus.EventBus
	Hasher         password.Hasher
	Tokenizer      token.Tokenizer
	V              validator.Validator
	Cfg            *config.Config
	TokenBlacklist domain.TokenBlacklist // optional; a new Redis adapter is created if nil
}

// NewModule builds and wires the auth module.
func NewModule(deps Deps, userRepo identitydomain.UserRepository) *Module {
	sessionRepo := persistence.NewPostgresSessionRepository(deps.DB, deps.Tx)
	sessionCache := authcache.NewRedisSessionCache(deps.Cache)

	tokenBlacklist := deps.TokenBlacklist
	if tokenBlacklist == nil {
		tokenBlacklist = authcache.NewRedisTokenBlacklist(deps.Cache)
	}

	loginCfg := commands.LoginConfig{
		AccessTTL:        deps.Cfg.JWTAccessTTL,
		RefreshTTL:       deps.Cfg.JWTRefreshTTL,
		MaxConcurrent:    deps.Cfg.SessionMaxConcurrent,
	}
	refreshCfg := commands.RefreshConfig{
		AccessTTL:  deps.Cfg.JWTAccessTTL,
		RefreshTTL: deps.Cfg.JWTRefreshTTL,
	}

	login := commands.NewLogin(userRepo, sessionRepo, sessionCache, tokenBlacklist, deps.Tokenizer, deps.Hasher, deps.Bus, deps.V, loginCfg)
	logout := commands.NewLogout(sessionRepo, sessionCache, tokenBlacklist, deps.Bus, deps.V)
	refresh := commands.NewRefresh(userRepo, sessionRepo, sessionCache, tokenBlacklist, deps.Tokenizer, deps.V, refreshCfg)
	revokeSession := commands.NewRevokeSession(sessionRepo, sessionCache, deps.Bus, deps.V)
	revokeAll := commands.NewRevokeAllSessions(sessionRepo, sessionCache, deps.Bus, deps.V)
	listSessions := queries.NewListSessions(sessionRepo, deps.V)

	authMW := authHTTP.NewAuthMiddleware(deps.Tokenizer, tokenBlacklist)
	handler := authHTTP.NewAuthHandler(login, logout, refresh, revokeSession, revokeAll, listSessions, deps.V)

	return &Module{
		authRouter: authHTTP.NewAuthRouter(handler, authMW),
		authMW:     authMW,
	}
}

// AuthRouter returns the chi router for auth endpoints.
func (m *Module) AuthRouter() chi.Router {
	return m.authRouter
}

// AuthMiddleware returns the authentication middleware so other modules can
// protect their routes without depending on auth's internal wiring.
func (m *Module) AuthMiddleware() func(http.Handler) http.Handler {
	return m.authMW.Authenticate
}
