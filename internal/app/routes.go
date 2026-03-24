package app

import (
	"net/http"

	authadapter "github.com/go-sum/auth/adapters/echocomponentry"
	"github.com/go-sum/forge/internal/adapters"
	"github.com/go-sum/forge/internal/handler"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes binds the application's concrete handlers to their URL paths.
// This is the single source of truth for HTTP route registration.
func RegisterRoutes(c *Container, h *handler.Handler, authH *authadapter.Handler) error {
	users := adapters.NewAuthUserReader(c.Repos.User)
	e := c.Web

	authKeys := authadapter.ContextKeys{
		UserID:      c.Config.App.Keys.UserID,
		UserRole:    c.Config.App.Keys.UserRole,
		DisplayName: c.Config.App.Keys.DisplayName,
	}
	e.Use(authadapter.LoadSession(c.Sessions, authKeys))

	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/health", Name: "health.show", Handler: h.HealthCheck})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: h.Home})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: authH.SigninPage})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signup", Name: "signup.get", Handler: authH.SignupPage})

	publicMutations := e.Group("")
	publicMutations.Use(
		appserver.ProtectBrowserMutation(c.Config),
		appserver.RateLimitMiddleware(c.Config, "auth"),
	)
	route.Add(publicMutations, echo.Route{Method: http.MethodPost, Path: "/signin", Name: "signin.post", Handler: authH.Signin})
	route.Add(publicMutations, echo.Route{Method: http.MethodPost, Path: "/signup", Name: "signup.post", Handler: authH.Signup})

	protected := e.Group("")
	protected.Use(
		authadapter.RequireAuthPath(func() string {
			return route.Reverse(e.Router().Routes(), "signin.get")
		}, authKeys),
		authadapter.LoadUserContext(users, authKeys),
	)
	route.Add(protected, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "components.list", Handler: h.ComponentExamples})

	protectedMutations := protected.Group("")
	protectedMutations.Use(appserver.ProtectBrowserMutation(c.Config))
	route.Add(protectedMutations, echo.Route{Method: http.MethodPost, Path: "/signout", Name: "signout.post", Handler: authH.Signout})

	usersGroup := protected.Group("/users")
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "", Name: "user.list", Handler: h.UserList})
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "/:id/edit", Name: "user.edit", Handler: h.UserEditForm})
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "user.row", Handler: h.UserRow})

	usersMutations := protectedMutations.Group("/users")
	route.Add(usersMutations, echo.Route{Method: http.MethodPut, Path: "/:id", Name: "user.update", Handler: h.UserUpdate})
	route.Add(usersMutations, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "user.delete", Handler: h.UserDelete})

	return nil
}
