// Package app provides the application composition root and service container.
//
// Container owns infrastructure wiring and domain composition; App owns the
// runnable HTTP application built from that container.
package app

import (
	"log/slog"

	authservice "github.com/go-sum/auth/service"
	"github.com/go-sum/auth/session"
	"github.com/go-sum/componentry/assetconfig"
	"github.com/go-sum/componentry/assets"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/repository"
	"github.com/go-sum/forge/internal/service"
	pkgserver "github.com/go-sum/server"
	"github.com/go-sum/server/database"
	"github.com/go-sum/server/validate"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Container holds all application services. Fields are populated by NewContainer
// in dependency order and are safe to read concurrently after construction.
type Container struct {
	Config       *config.Config
	AssetPaths   assetconfig.Paths
	DB           *pgxpool.Pool
	Assets       *assets.Assets
	Web          *echo.Echo
	ServerConfig pkgserver.Config
	PublicDir    string // filesystem path to built public assets, e.g. "public"
	Sessions     *session.SessionManager
	Validator    *validate.Validator
	Repos        *repository.Repositories
	Services     *service.Services
	AuthService  *authservice.AuthService
}

// NewContainer initialises all services in dependency order.
// Panics on any fatal startup failure; these are non-recoverable at init time.
func NewContainer() *Container {
	c := &Container{}
	c.initConfig()
	c.initLogger()
	c.initAssets()
	c.initDatabase()
	c.initWeb()
	c.initAuth()
	c.initValidator()
	c.initRepos()
	c.initServices()
	return c
}

// Shutdown gracefully tears down all services held by the container.
func (c *Container) Shutdown() {
	database.Close(c.DB)
	slog.Info("container shutdown complete")
}
