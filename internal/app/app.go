package app

import (
	"context"
	"fmt"

	authadapter "github.com/go-sum/auth/adapters/echocomponentry"
	"github.com/go-sum/forge/internal/handler"
	"github.com/go-sum/forge/internal/view"
	pkgserver "github.com/go-sum/server"
	"github.com/go-sum/server/database"
	route "github.com/go-sum/server/route"

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
			CSRFField:      c.ServerConfig.CSRFCookieName,
			LoginPathFn:    func() string { return route.Reverse(c.Web.Router().Routes(), "session.new") },
			RegisterPathFn: func() string { return route.Reverse(c.Web.Router().Routes(), "registration.new") },
			HomePathFn:     func() string { return route.Reverse(c.Web.Router().Routes(), "home.show") },
			RequestFn: func(ec *echo.Context) authadapter.Request {
				req := view.NewRequest(ec, c.Config.Nav)
				return authadapter.Request{
					CSRFToken: req.CSRFToken,
					PageFn:    req.Page,
				}
			},
		},
	)

	c.Web.Static(c.ServerConfig.PublicPrefix, c.PublicDir)
	if err := RegisterRoutes(c, h, authH); err != nil {
		panic(fmt.Sprintf("routes: %v", err))
	}
	return &App{container: c}
}

// Start starts the HTTP server.
func (a *App) Start() error {
	return pkgserver.Start(a.container.Web, a.container.ServerConfig)
}

// Shutdown releases application resources.
func (a *App) Shutdown() {
	a.container.Shutdown()
}
