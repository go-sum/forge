package app

import (
	"context"
	"fmt"

	auth "github.com/go-sum/auth"
	authadapter "github.com/go-sum/forge/internal/adapters/auth"
	"github.com/go-sum/forge/internal/handler"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/server"
	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

// App owns the fully wired application and its lifecycle.
type App struct {
	container *Container
}

// New initializes the application container and registers all HTTP routes.
// version is the build-time application version (may be empty in dev).
func New(version string) *App {
	c := NewContainer()
	availabilityH := handler.NewAvailability(c.checkHealth(), c.StartupError, version)

	if c.StartupError != nil {
		if err := RegisterStartupRoutes(c, availabilityH); err != nil {
			panic(fmt.Sprintf("routes: %v", err))
		}
		return &App{container: c}
	}

	h := handler.New(
		c.Config,
		func() echo.Routes { return c.Web.Router().Routes() },
		c.Services,
		c.Validator,
	)

	authH := auth.NewHandler(
		c.AuthService,
		auth.HandlerConfig{
			Sessions:           &authadapter.SessionManagerAdapter{Mgr: c.Sessions},
			Forms:              &authadapter.FormParserAdapter{V: c.Validator},
			Flash:              &authadapter.FlashAdapter{},
			Redirect:           &authadapter.RedirectAdapter{},
			Pages:              authadapter.NewRenderer(),
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
			RequestFn: func(ec *echo.Context) auth.Request {
				req := view.NewRequest(ec, c.Config)
				return auth.Request{
					CSRFToken: req.CSRFToken,
					PageFn:    req.Page,
				}
			},
		},
	)

	if err := RegisterRoutes(c, h, availabilityH, authH); err != nil {
		panic(fmt.Sprintf("routes: %v", err))
	}
	return &App{container: c}
}

// Start launches all registered background services and the HTTP server.
func (a *App) Start() error {
	a.container.StartBackground(context.Background())
	return server.Start(a.container.Web, a.container.ServerConfig)
}

// Shutdown releases application resources.
func (a *App) Shutdown() {
	a.container.Shutdown()
}
