// Package app provides the application composition root and service container.
//
// Container owns infrastructure wiring and domain composition; App owns the
// runnable HTTP application built from that container.
package app

import (
	"log/slog"

	auth "github.com/go-sum/auth"
	"github.com/go-sum/kv"
	"github.com/go-sum/session"
	"github.com/go-sum/componentry/assets"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/repository"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/go-sum/forge/internal/service"
	"github.com/go-sum/send"
	"github.com/go-sum/server"
	"github.com/go-sum/server/validate"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Container holds all application services. Fields are populated by NewContainer
// in dependency order and are safe to read concurrently after construction.
type Container struct {
	Config       *config.Config
	PublicPrefix string // URL prefix for static assets, e.g. "/public"
	DB           *pgxpool.Pool
	Assets       *assets.Assets
	Web          *echo.Echo
	ServerConfig server.Config
	RateLimiters *appserver.RateLimiters
	PublicDir    string // filesystem path to built public assets, e.g. "public"
	Sessions     session.Manager
	KV           kv.Store
	Validator    *validate.Validator
	Repos        *repository.Repositories
	Services     *service.Services
	AuthService  auth.Service
	Sender       send.Sender
}

// NewContainer initialises all services in dependency order.
// Panics on any fatal startup failure; these are non-recoverable at init time.
func NewContainer() *Container {
	c := &Container{}
	c.initConfig()
	c.initLogger()
	c.initSender()
	c.initAssets()
	c.initDatabase()
	c.initKV()
	c.initWeb()
	c.initAuth()
	c.initValidator()
	c.initRepos()
	c.initServices()
	return c
}

// Shutdown gracefully tears down all services held by the container.
func (c *Container) Shutdown() {
	if c.KV != nil {
		if err := c.KV.Close(); err != nil {
			slog.Error("kv close error", "error", err)
		}
	}
	c.DB.Close()
	slog.Info("container shutdown complete")
}
