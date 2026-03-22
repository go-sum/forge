package app

import (
	"context"

	authadapter "github.com/y-goweb/auth/adapters/echocomponentry"
	"github.com/y-goweb/foundry/internal/adapters"
	"github.com/y-goweb/foundry/internal/handler"
	"github.com/y-goweb/foundry/internal/routes"
	"github.com/y-goweb/foundry/internal/server"
	"github.com/y-goweb/foundry/internal/view"
	pkgserver "github.com/y-goweb/server"
	"github.com/y-goweb/server/database"

	"github.com/labstack/echo/v5"
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
		c.Validator,
		func(ctx context.Context) error { return database.CheckHealth(ctx, c.DB) },
		c.Config.Nav,
	)

	authH := authadapter.New(
		c.AuthService,
		c.Sessions,
		c.Validator,
		authadapter.Config{
			LoginPath:    routes.Login,
			RegisterPath: routes.Register,
			HomePath:     routes.Home,
			CSRFField:    c.ServerConfig.CSRFCookieName,
			RequestFn: func(ec *echo.Context) authadapter.Request {
				req := view.NewRequest(ec, c.Config.Nav)
				return authadapter.Request{
					CSRFToken: req.CSRFToken,
					PageFn:    req.Page,
				}
			},
		},
	)

	server.RegisterRoutes(c.Web, h, authH, c.Sessions, adapters.NewAuthUserReader(c.Repos.User), c.ServerConfig.PublicPrefix, c.PublicDir)
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
