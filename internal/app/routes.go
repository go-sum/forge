package app

import (
	"net/http"

	authadapter "github.com/go-sum/auth/adapters/echocomponentry"
	"github.com/go-sum/forge/internal/adapters"
	"github.com/go-sum/forge/internal/handler"
	route "github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes binds the application's concrete handlers to their URL paths.
// This is the single source of truth for HTTP route registration.
func RegisterRoutes(c *Container, h *handler.Handler, authH *authadapter.Handler) error {
	users := adapters.NewAuthUserReader(c.Repos.User)
	e := c.Web

	e.Use(authadapter.LoadSession(c.Sessions))

	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/health", Name: "health.show", Handler: h.HealthCheck})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: h.Home})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/login", Name: "session.new", Handler: authH.LoginPage})
	route.Add(e, echo.Route{Method: http.MethodPost, Path: "/login", Name: "session.create", Handler: authH.Login})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/register", Name: "registration.new", Handler: authH.RegisterPage})
	route.Add(e, echo.Route{Method: http.MethodPost, Path: "/register", Name: "registration.create", Handler: authH.Register})

	protected := e.Group("")
	protected.Use(
		authadapter.RequireAuthPath(func() string {
			return route.Reverse(e.Router().Routes(), "session.new")
		}),
		authadapter.LoadUserContext(users),
	)
	route.Add(protected, echo.Route{Method: http.MethodPost, Path: "/logout", Name: "session.delete", Handler: authH.Logout})
	route.Add(protected, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "component-example.list", Handler: h.ComponentExamples})

	usersGroup := protected.Group("/users")
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "", Name: "user.list", Handler: h.UserList})
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "/:id/edit", Name: "user.edit", Handler: h.UserEditForm})
	route.Add(usersGroup, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "user.row", Handler: h.UserRow})
	route.Add(usersGroup, echo.Route{Method: http.MethodPut, Path: "/:id", Name: "user.update", Handler: h.UserUpdate})
	route.Add(usersGroup, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "user.delete", Handler: h.UserDelete})

	return nil
}
