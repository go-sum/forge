package app

import (
	"context"
	"fmt"

	"github.com/go-sum/forge/internal/adapters/authui"
	"github.com/go-sum/forge/internal/handler"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/server"
	"github.com/go-sum/server/database"
	"github.com/go-sum/server/route"

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
		c.Config,
		func() echo.Routes { return c.Web.Router().Routes() },
	)

	authH := authui.New(
		c.AuthService,
		c.Sessions,
		c.Validator,
		authui.Config{
			CSRFField:          c.Config.App.Security.CSRF.FormField,
			SigninPathFn:       func() string { return route.Reverse(c.Web.Router().Routes(), "signin.get") },
			SignupPathFn:       func() string { return route.Reverse(c.Web.Router().Routes(), "signup.get") },
			VerifyPathFn:       func() string { return route.Reverse(c.Web.Router().Routes(), "verify.get") },
			VerifyResendPathFn: func() string { return route.Reverse(c.Web.Router().Routes(), "verify.resend.post") },
			VerifyURLFn: func() string {
				return c.Config.App.Security.ExternalOrigin + route.Reverse(c.Web.Router().Routes(), "verify.get")
			},
			EmailChangeFn: func() string { return route.Reverse(c.Web.Router().Routes(), "account.email.get") },
			HomePathFn:    func() string { return route.Reverse(c.Web.Router().Routes(), "home.show") },
			RequestFn: func(ec *echo.Context) authui.Request {
				req := view.NewRequest(ec, c.Config)
				return authui.Request{
					CSRFToken: req.CSRFToken,
					PageFn:    req.Page,
				}
			},
		},
	)

	c.Web.Static(c.AssetPaths.URLPrefix(), c.PublicDir)
	if err := RegisterRoutes(c, h, authH); err != nil {
		panic(fmt.Sprintf("routes: %v", err))
	}
	return &App{container: c}
}

// Start starts the HTTP server.
func (a *App) Start() error {
	return server.Start(a.container.Web, a.container.ServerConfig)
}

// Shutdown releases application resources.
func (a *App) Shutdown() {
	a.container.Shutdown()
}
