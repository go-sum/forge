package app

import (
	"net/http"

	authadapter "github.com/go-sum/auth/adapters/echocomponentry"
	"github.com/go-sum/forge/internal/adapters"
	"github.com/go-sum/forge/internal/handler"
	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes binds the application's concrete handlers to their URL paths.
// This is the single source of truth for HTTP route registration.
func RegisterRoutes(c *Container, h *handler.Handler, authH *authadapter.Handler) error {
	users := adapters.NewAuthUserReader(c.Repos.User)
	e := c.Web

	authKeys := authadapter.ContextKeys{
		UserID:      c.Config.Keys.UserID,
		UserRole:    c.Config.Keys.UserRole,
		DisplayName: c.Config.Keys.DisplayName,
	}
	e.Use(authadapter.LoadSession(c.Sessions, authKeys))

	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/health", Name: "health.show", Handler: h.HealthCheck})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: h.Home})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: authH.SigninPage})
	route.Add(e, echo.Route{Method: http.MethodPost, Path: "/signin", Name: "signin.post", Handler: authH.Signin})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signup", Name: "signup.get", Handler: authH.SignupPage})
	route.Add(e, echo.Route{Method: http.MethodPost, Path: "/signup", Name: "signup.post", Handler: authH.Signup})

	protected := e.Group("")
	protected.Use(
		authadapter.RequireAuthPath(func() string {
			return route.Reverse(e.Router().Routes(), "signin.get")
		}, authKeys),
		authadapter.LoadUserContext(users, authKeys),
	)
	route.Add(protected, echo.Route{Method: http.MethodPost, Path: "/signout", Name: "signout.post", Handler: authH.Signout})
	route.Add(protected, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "components.list", Handler: h.ComponentExamples})

	usersGroup := protected.Group("/users")
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "", Name: "user.list", Handler: h.UserList})
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "/:id/edit", Name: "user.edit", Handler: h.UserEditForm})
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "user.row", Handler: h.UserRow})
	route.Add(usersGroup, echo.Route{Method: http.MethodPut, Path: "/:id", Name: "user.update", Handler: h.UserUpdate})
	route.Add(usersGroup, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "user.delete", Handler: h.UserDelete})

	return nil
}
