// Package http wires the chi router with platform middleware and starts the server.
package http

import (
	"fmt"
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

// Start runs the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.cfg.HTTPHost, s.cfg.HTTPPort)
	s.log.Info("starting http server", "address", addr)
	return http.ListenAndServe(addr, s.router)
}
