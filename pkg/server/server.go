// Package server provides a generic Echo v5 instance factory and lifecycle manager.
// New creates the bare Echo instance; Start handles graceful shutdown via OS signals.
// Application-specific concerns (middleware, routing) live in internal/server.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
)

type Config struct {
	Host            string
	Port            string
	Debug           bool
	GracefulTimeout time.Duration
	CookieSecure    bool
	CSP             string
	CSRFCookieName  string
	PublicPrefix    string // URL prefix for public assets, e.g. "/public"
}

// New creates a bare Echo v5 instance with only the error handler configured.
// Call internal/server.Setup to wire the application's middleware stack.
func New(cfg Config) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = echo.DefaultHTTPErrorHandler(cfg.Debug)
	return e
}

// Start begins listening and blocks until a shutdown signal is received.
// It returns an error rather than calling os.Exit so that deferred cleanup
// in main (e.g. database.Close) runs on shutdown.
func Start(e *echo.Echo, cfg Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sc := echo.StartConfig{
		Address:         cfg.Host + ":" + cfg.Port,
		GracefulTimeout: cfg.GracefulTimeout,
		OnShutdownError: func(err error) {
			slog.Error("server forced to shutdown", "error", err)
		},
	}

	slog.Info("server starting", "address", sc.Address)
	if err := sc.Start(ctx, e); err != nil {
		return fmt.Errorf("server: %w", err)
	}
	slog.Info("server stopped")
	return nil
}
