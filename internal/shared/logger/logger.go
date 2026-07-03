// Package logger provides a tiny structured logger wrapper around the standard
// library log/slog.
//
// Why slog?
// - Added in Go 1.21, zero external dependency.
// - Structured JSON output in production, text output in development.
package logger

import (
	"log/slog"
	"os"
)

// Logger is a thin alias so the rest of the app does not depend directly on slog.
type Logger = *slog.Logger

// New creates a Logger.
func New(environment string) Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	var handler slog.Handler
	if environment == "development" || environment == "local" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
