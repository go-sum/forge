package app

import (
	"context"

	"starter/internal/handler"
	"starter/internal/server"
	"starter/pkg/database"
	pkgserver "starter/pkg/server"
)

// App owns the fully wired application and its lifecycle.
type App struct {
	container *Container
}

// New initializes the application container and registers all HTTP routes.
func New() *App {
	c := NewContainer()
	h := handler.New(
		c.Services,
		c.Sessions,
		c.Validator,
		func(ctx context.Context) error { return database.CheckHealth(ctx, c.DB) },
		c.Config.Nav,
	)
	server.RegisterRoutes(c.Web, h, c.Sessions, c.Services.User, c.ServerConfig.PublicPrefix, c.PublicDir)
	return &App{container: c}
}

// Run starts the HTTP server.
func (a *App) Run() error {
	return pkgserver.Start(a.container.Web, a.container.ServerConfig)
}

// Shutdown releases application resources.
func (a *App) Shutdown() {
	a.container.Shutdown()
}
