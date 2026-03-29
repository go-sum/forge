// Package server provides an Echo v5 instance factory and lifecycle manager.
// NewWithConfig creates an Echo instance with construction-time hooks;
// Start handles graceful shutdown via OS signals.
// Application-specific concerns (middleware, routing) live in internal/server.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
)

type Config struct {
	Host            string
	Port            string
	GracefulTimeout time.Duration
	// BeforeServeFunc is called with the underlying *http.Server before it
	// starts accepting connections. Use to set ReadTimeout, WriteTimeout,
	// IdleTimeout, MaxHeaderBytes, etc. Optional.
	BeforeServeFunc func(*http.Server) error
	// ListenerAddrFunc is called with the resolved net.Addr after the listener
	// is bound. Use to discover the actual port when Port is "0". Optional.
	ListenerAddrFunc func(net.Addr)
}

// NewWithConfig creates an Echo v5 instance with construction-time configuration.
// Pass echo.Config{HTTPErrorHandler: ...} to install the error handler at
// construction time rather than post-construction via field assignment.
func NewWithConfig(cfg echo.Config) *echo.Echo {
	return echo.NewWithConfig(cfg)
}

// Start begins listening and blocks until a shutdown signal is received.
// It returns an error rather than calling os.Exit so that deferred cleanup
// in main (e.g. pool.Close) runs on shutdown.
func Start(e *echo.Echo, cfg Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sc := echo.StartConfig{
		Address:          cfg.Host + ":" + cfg.Port,
		GracefulTimeout:  cfg.GracefulTimeout,
		BeforeServeFunc:  cfg.BeforeServeFunc,
		ListenerAddrFunc: cfg.ListenerAddrFunc,
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
