package app

import (
	"context"
	"fmt"

	auth "github.com/go-sum/auth"
	authsvc "github.com/go-sum/auth/service"
	authadapter "github.com/go-sum/forge/internal/adapters/auth"
	"github.com/go-sum/forge/internal/features/availability"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/server"
	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

// App owns the fully wired application and its lifecycle.
type App struct {
	runtime *Runtime
}

// New initializes the application runtime and registers all HTTP routes.
// version is the build-time application version (may be empty in dev).
func New(version string) *App {
	r := NewRuntime()
	r.Config.App.Version = version
	availabilityH := availability.NewHandler(r.checkHealth(), r.StartupError, version)

	if r.StartupError != nil {
		if err := RegisterStartupRoutes(r, availabilityH); err != nil {
			panic(fmt.Sprintf("routes: %v", err))
		}
		return &App{runtime: r}
	}

	resolve := route.NewResolver(func() echo.Routes { return r.Web.Router().Routes() })

	requestFn := func(ec *echo.Context) auth.Request {
		req := view.NewRequest(ec, r.Config)
		return auth.Request{
			CSRFToken:     req.CSRFToken,
			CSRFFieldName: req.CSRFFieldName,
			Partial:       req.IsPartial(),
			State:         req,
			PageFn:        req.Page,
		}
	}

	authH := auth.NewHandler(
		r.AuthService,
		auth.HandlerConfig{
			Sessions:           &authadapter.SessionManagerAdapter{Mgr: r.Sessions},
			Forms:              &authadapter.FormParserAdapter{V: r.Validator},
			Flash:              &authadapter.FlashAdapter{},
			Redirect:           &authadapter.RedirectAdapter{},
			Pages:              authadapter.NewRenderer(),
			CSRFField:          r.Config.Security.CSRF.FormField,
			SigninPathFn:       resolve.Path("signin.get"),
			SignupPathFn:       resolve.Path("signup.get"),
			VerifyPathFn:       resolve.Path("verify.get"),
			VerifyResendPathFn: resolve.Path("verify.resend.post"),
			VerifyURLFn:        resolve.URL(r.Config.Security.ExternalOrigin, "verify.get"),
			EmailChangeFn:      resolve.Path("account.email.get"),
			HomePathFn:         resolve.Path("home.show"),
			RequestFn:          requestFn,
		},
	)

	adminH := auth.NewAdminHandler(
		authsvc.NewAdminService(r.AuthStore),
		auth.AdminHandlerConfig{
			Forms:      &authadapter.FormParserAdapter{V: r.Validator},
			Redirect:   &authadapter.RedirectAdapter{},
			Pages:      authadapter.NewRenderer(),
			HomePathFn: resolve.Path("home.show"),
			RequestFn:  requestFn,
		},
	)

	if err := RegisterRoutes(r, availabilityH, authH, adminH); err != nil {
		panic(fmt.Sprintf("routes: %v", err))
	}
	return &App{runtime: r}
}

// Start launches all registered background services and the HTTP server.
func (a *App) Start() error {
	a.runtime.StartBackground(context.Background())
	return server.Start(a.runtime.Web, a.runtime.ServerConfig)
}

// Shutdown releases application resources.
func (a *App) Shutdown() {
	a.runtime.Shutdown()
}
