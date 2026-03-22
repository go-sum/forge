package server

import (
	authadapter "github.com/go-sum/auth/adapters/echocomponentry"
	authrepo "github.com/go-sum/auth/repository"
	"github.com/go-sum/auth/session"
	"github.com/go-sum/forge/internal/routes"

	"github.com/labstack/echo/v5"
)

// Handlers is satisfied by *handler.Handler through Go's structural typing.
// Adding a route requires: a new constant in internal/routes, a new method here,
// and a new registration line in RegisterRoutes — all three in sync.
type Handlers interface {
	HealthCheck(*echo.Context) error
	Home(*echo.Context) error
	ComponentExamples(*echo.Context) error
	UserList(*echo.Context) error
	UserEditForm(*echo.Context) error
	UserRow(*echo.Context) error
	UserUpdate(*echo.Context) error
	UserDelete(*echo.Context) error
}

// AuthHandlers is satisfied by *authhandler.Handler through Go's structural typing.
type AuthHandlers interface {
	LoginPage(*echo.Context) error
	Login(*echo.Context) error
	RegisterPage(*echo.Context) error
	Register(*echo.Context) error
	Logout(*echo.Context) error
}

// RegisterRoutes is the single source of truth for route paths, HTTP methods,
// handler assignments, and per-group middleware. Adding a route touches only
// this file (plus the handler implementation in internal/handler/).
func RegisterRoutes(
	e *echo.Echo,
	h Handlers,
	authH AuthHandlers,
	sessions *session.SessionManager,
	users authrepo.UserReader,
	publicPrefix, publicDir string,
) {
	e.Static(publicPrefix, publicDir)
	e.Use(authadapter.LoadSession(sessions))

	e.GET(routes.Health, h.HealthCheck)
	e.GET(routes.Home, h.Home)

	e.GET(routes.Login, authH.LoginPage)
	e.POST(routes.Login, authH.Login)
	e.GET(routes.Register, authH.RegisterPage)
	e.POST(routes.Register, authH.Register)

	protected := e.Group("")
	protected.Use(authadapter.RequireAuth(routes.Login))
	protected.Use(authadapter.LoadUserContext(users))
	protected.POST(routes.Logout, authH.Logout)
	protected.GET(routes.Components, h.ComponentExamples)
	protected.GET(routes.Users, h.UserList)
	protected.GET(routes.UserEdit, h.UserEditForm)
	protected.GET(routes.UserRow, h.UserRow)
	protected.PUT(routes.UserByID, h.UserUpdate)
	protected.DELETE(routes.UserByID, h.UserDelete)
}
