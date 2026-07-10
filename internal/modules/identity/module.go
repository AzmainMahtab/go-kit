// Package identity wires the identity bounded context.
package identity

import (
	"net/http"

	"github.com/example/go-kit/internal/modules/identity/application/commands"
	"github.com/example/go-kit/internal/modules/identity/application/queries"
	"github.com/example/go-kit/internal/modules/identity/domain"
	"github.com/example/go-kit/internal/modules/identity/infrastructure/persistence"
	identityHTTP "github.com/example/go-kit/internal/modules/identity/presentation/http"
	"github.com/example/go-kit/internal/platform/database"
	"github.com/example/go-kit/internal/shared/eventbus"
	"github.com/example/go-kit/internal/shared/password"
	"github.com/example/go-kit/internal/shared/validator"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// Module is the public API of the identity bounded context.
type Module struct {
	userRouter chi.Router
}

// Deps contains the external dependencies identity needs.
type Deps struct {
	DB     *sqlx.DB
	Tx     *database.TxManager
	Bus    eventbus.EventBus
	Hasher password.Hasher
	V      validator.Validator
	// AuthMW protects identity routes. It is provided by the auth module.
	AuthMW func(http.Handler) http.Handler
}

// NewModule builds and wires the identity module.
func NewModule(deps Deps) *Module {
	userRepo := persistence.NewPostgresUserRepository(deps.DB, deps.Tx)

	register := commands.NewRegisterUser(userRepo, deps.Bus, deps.Hasher, deps.V)
	update := commands.NewUpdateUser(userRepo, deps.Bus, deps.V)
	getUser := queries.NewGetUser(userRepo, deps.V)
	listUsers := queries.NewListUsers(userRepo, deps.V)

	handler := identityHTTP.NewUserHandler(register, update, getUser, listUsers, deps.V)

	return &Module{
		userRouter: identityHTTP.NewUserRouter(handler, deps.AuthMW),
	}
}

// UserRouter returns the chi router for user endpoints.
func (m *Module) UserRouter() chi.Router {
	return m.userRouter
}

// UserRepository exposes the repository for cross-module ACL use.
func (m *Module) UserRepository(db *sqlx.DB, tx *database.TxManager) domain.UserRepository {
	return persistence.NewPostgresUserRepository(db, tx)
}
