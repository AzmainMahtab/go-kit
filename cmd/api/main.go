// Package main is the application entrypoint.
//
// Think of main.go as the composition root: it builds every concrete adapter
// and injects them into modules. No business logic lives here.
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	_ "github.com/elite4print/elite4print-go/docs" // swagger docs
	"github.com/elite4print/elite4print-go/internal/modules/auth"
	authcache "github.com/elite4print/elite4print-go/internal/modules/auth/infrastructure/cache"
	authHTTP "github.com/elite4print/elite4print-go/internal/modules/auth/presentation/http"
	"github.com/elite4print/elite4print-go/internal/modules/identity"
	"github.com/elite4print/elite4print-go/internal/platform/http/responses"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/elite4print/elite4print-go/internal/platform/cache"
	"github.com/elite4print/elite4print-go/internal/platform/config"
	"github.com/elite4print/elite4print-go/internal/platform/database"
	platformhttp "github.com/elite4print/elite4print-go/internal/platform/http"
	"github.com/elite4print/elite4print-go/internal/platform/http/middleware"
	"github.com/elite4print/elite4print-go/internal/shared/eventbus"
	"github.com/elite4print/elite4print-go/internal/shared/logger"
	"github.com/elite4print/elite4print-go/internal/shared/password"
	"github.com/elite4print/elite4print-go/internal/shared/token"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
)

// @title Elite4Print Go API
// @version 1.0
// @description Modular monolith backend for Elite4Print.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log := logger.New(cfg.Environment)

	db, err := database.NewPool(cfg)
	if err != nil {
		log.Error("failed to connect to database", slog.Any("error", err))
		return
	}
	defer db.Close()

	txManager := database.NewTxManager(db)

	redisCache := cache.NewRedis(cfg)
	if err := redisCache.Ping(context.Background()); err != nil {
		log.Warn("redis unreachable", slog.Any("error", err))
	}
	defer redisCache.Close()

	// Use in-memory event bus for the starter set. Swap for NATS JetStream when
	// you are ready to run background workers across instances.
	var bus eventbus.EventBus = eventbus.NewInMemory()

	v := validator.New()
	hasher := password.NewArgon2id()
	tokenizer := token.NewJWT(cfg.JWTSecretKey)

	rateLimiter := middleware.NewRateLimiter(redisCache, cfg.RateLimitRPS, cfg.RateLimitBurst)
	server := platformhttp.NewServer(cfg, log, rateLimiter)

	// Shared auth middleware is built before modules so it can protect identity
	// routes as well as auth routes.
	tokenBlacklist := authcache.NewRedisTokenBlacklist(redisCache)
	authMW := authHTTP.NewAuthMiddleware(tokenizer, tokenBlacklist)

	// Identity module.
	identityModule := identity.NewModule(identity.Deps{
		DB:     db,
		Tx:     txManager,
		Bus:    bus,
		Hasher: hasher,
		V:      v,
		AuthMW: authMW.Authenticate,
	})
	userRepo := identityModule.UserRepository(db, txManager)

	// Auth module.
	authModule := auth.NewModule(auth.Deps{
		DB:             db,
		Tx:             txManager,
		Cache:          redisCache,
		Bus:            bus,
		Hasher:         hasher,
		Tokenizer:      tokenizer,
		V:              v,
		Cfg:            cfg,
		TokenBlacklist: tokenBlacklist,
	}, userRepo)

	server.Router().Mount("/api/v1/users", identityModule.UserRouter())
	server.Router().Mount("/api/v1/auth", authModule.AuthRouter())

	// Health check.
	server.Router().Get("/health", func(w http.ResponseWriter, r *http.Request) {
		responses.OK(w, map[string]string{"status": "ok"})
	})

	// Swagger UI.
	server.Router().Get("/swagger/*", httpSwagger.WrapHandler)

	if err := server.Start(); err != nil {
		log.Error("server crashed", slog.Any("error", err))
	}
}
