// Package http wires the chi router with platform middleware and starts the server.
package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/elite4print/elite4print-go/internal/platform/config"
	"github.com/elite4print/elite4print-go/internal/platform/http/middleware"
	"github.com/elite4print/elite4print-go/internal/shared/logger"
	"github.com/go-chi/chi/v5"
	chiMW "github.com/go-chi/chi/v5/middleware"
)

// Server wraps the chi router and configuration.
type Server struct {
	cfg    *config.Config
	router *chi.Mux
	log    logger.Logger
	http   *http.Server
}

// NewServer creates a server with common middleware mounted.
func NewServer(cfg *config.Config, log logger.Logger, rateLimiter *middleware.RateLimiter) *Server {
	r := chi.NewRouter()

	// Built-in chi middleware.
	r.Use(chiMW.RealIP)
	r.Use(chiMW.RequestSize(10 << 20)) // 10 MB max body
	r.Use(chiMW.Timeout(30 * 1000))    // 30s timeout

	// Custom platform middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(log))
	r.Use(middleware.Recovery(log))
	r.Use(rateLimiter.Handler)

	// CORS (adjust origins in production).
	r.Use(chiMW.SetHeader("Access-Control-Allow-Origin", "*"))
	r.Use(chiMW.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS"))
	r.Use(chiMW.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID"))

	return &Server{cfg: cfg, router: r, log: log}
}

// Router returns the underlying chi router so modules can mount sub-routers.
func (s *Server) Router() *chi.Mux {
	return s.router
}

// ListenAndServe starts the HTTP server in a non-blocking way.
// The caller is responsible for waiting on a shutdown signal and calling
// Shutdown to drain active connections.
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%s", s.cfg.HTTPHost, s.cfg.HTTPPort)
	s.http = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	s.log.Info("starting http server", "address", addr)
	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
	s.log.Info("shutting down http server")
	return s.http.Shutdown(ctx)
}
