// Package services provides the application service container — a single
// composition root that owns and initialises all infrastructure services in
// dependency order. Domain layers (Repos, Services, Handler) are added in T1202.
package services

import (
	"log/slog"

	"starter/config"
	"starter/pkg/assetconfig"
	"starter/pkg/assets"
	"starter/pkg/database"
	pkgserver "starter/pkg/server"

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
	return c
}

// Shutdown gracefully tears down all services held by the container.
func (c *Container) Shutdown() {
	database.Close(c.DB)
	slog.Info("container shutdown complete")
}
