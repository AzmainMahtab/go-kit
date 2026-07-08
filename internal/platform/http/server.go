// Package http wires the chi router with platform middleware and runs the
// HTTP server. It keeps cmd/api/main.go small by owning route mounting,
// lifecycle, and graceful shutdown.
package http

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/elite4print/elite4print-go/internal/platform/config"
	"github.com/elite4print/elite4print-go/internal/platform/health"
	"github.com/elite4print/elite4print-go/internal/platform/http/middleware"
	"github.com/elite4print/elite4print-go/internal/platform/http/responses"
	"github.com/elite4print/elite4print-go/internal/shared/logger"
	"github.com/go-chi/chi/v5"
	chiMW "github.com/go-chi/chi/v5/middleware"
)

// ServerDeps contains everything the HTTP server needs to mount routes and
// expose endpoints.
type ServerDeps struct {
	Config         *config.Config
	Log            logger.Logger
	RateLimiter    *middleware.RateLimiter
	AuthRouter     chi.Router
	IdentityRouter chi.Router
	HealthChecker  *health.Checker
	SwaggerHandler http.Handler
}

// Server wraps the chi router and the underlying http.Server.
type Server struct {
	cfg    *config.Config
	router *chi.Mux
	log    logger.Logger
	http   *http.Server
}

// NewServer creates and configures the HTTP server.
// It mounts all application routes so main stays a thin composition root.
func NewServer(deps ServerDeps) *Server {
	r := chi.NewRouter()

	// Built-in chi middleware.
	r.Use(chiMW.RealIP)
	r.Use(chiMW.RequestSize(10 << 20)) // 10 MB max body
	r.Use(chiMW.Timeout(30 * 1000))    // 30s handler timeout

	// Custom platform middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(deps.Log))
	r.Use(middleware.Recovery(deps.Log))
	r.Use(deps.RateLimiter.Handler)

	// CORS (adjust origins in production when credentials are used).
	r.Use(chiMW.SetHeader("Access-Control-Allow-Origin", "*"))
	r.Use(chiMW.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS"))
	r.Use(chiMW.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID"))

	// Application routes.
	r.Mount("/api/v1/auth", deps.AuthRouter)
	r.Mount("/api/v1/users", deps.IdentityRouter)

	// Health check.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		result := deps.HealthChecker.Check(r.Context())
		if result.Status == health.StatusDown {
			responses.JSON(w, http.StatusServiceUnavailable, responses.Error(http.StatusServiceUnavailable, "HEALTH_CHECK_FAILED", "one or more dependencies are unavailable"))
			return
		}
		responses.OK(w, result)
	})

	// Swagger UI.
	r.Get("/swagger/*", deps.SwaggerHandler.ServeHTTP)

	return &Server{cfg: deps.Config, router: r, log: deps.Log}
}

// Run starts the server and blocks until a SIGINT/SIGTERM is received.
// It then performs a graceful shutdown with a 30 second timeout.
func (s *Server) Run() error {
	if err := s.start(); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	s.log.Info("shutting down http server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.Shutdown(shutdownCtx)
}

func (s *Server) start() error {
	addr := fmt.Sprintf("%s:%s", s.cfg.HTTPHost, s.cfg.HTTPPort)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.http = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.log.Info("starting http server", "address", addr)
	go func() {
		if err := s.http.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.log.Error("http server error", slog.Any("error", err))
		}
	}()

	return nil
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}
