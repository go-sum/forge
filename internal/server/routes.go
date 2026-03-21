package server

import (
	custommw "starter/internal/middleware"
	"starter/internal/routes"
	"starter/internal/service"
	"starter/pkg/auth"

	"github.com/labstack/echo/v5"
)

// Handlers is satisfied by *handler.Handler through Go's structural typing.
// Adding a route requires: a new constant in internal/routes, a new method here,
// and a new registration line in RegisterRoutes — all three in sync.
type Handlers interface {
	HealthCheck(*echo.Context) error
	Home(*echo.Context) error
	LoginPage(*echo.Context) error
	Login(*echo.Context) error
	RegisterPage(*echo.Context) error
	Register(*echo.Context) error
	Logout(*echo.Context) error
	ComponentExamples(*echo.Context) error
	UserList(*echo.Context) error
	UserEditForm(*echo.Context) error
	UserRow(*echo.Context) error
	UserUpdate(*echo.Context) error
	UserDelete(*echo.Context) error
}

// RegisterRoutes is the single source of truth for route paths, HTTP methods,
// handler assignments, and per-group middleware. Adding a route touches only
// this file (plus the handler implementation in internal/handler/).
func RegisterRoutes(
	e *echo.Echo,
	h Handlers,
	sessions *auth.SessionManager,
	users *service.UserService,
	publicPrefix, publicDir string,
) {
	e.Static(publicPrefix, publicDir)
	e.Use(custommw.LoadSession(sessions))

	e.GET(routes.Health, h.HealthCheck)
	e.GET(routes.Home, h.Home)

	e.GET(routes.Login, h.LoginPage)
	e.POST(routes.Login, h.Login)
	e.GET(routes.Register, h.RegisterPage)
	e.POST(routes.Register, h.Register)

	protected := e.Group("")
	protected.Use(custommw.RequireAuth(routes.Login))
	protected.POST(routes.Logout, h.Logout)
	protected.GET(routes.Components, h.ComponentExamples)

	admin := protected.Group("")
	admin.Use(custommw.LoadUserRole(users))
	admin.Use(custommw.RequireAdmin())
	admin.GET(routes.Users, h.UserList)
	admin.GET(routes.UserEdit, h.UserEditForm)
	admin.GET(routes.UserRow, h.UserRow)
	admin.PUT(routes.UserByID, h.UserUpdate)
	admin.DELETE(routes.UserByID, h.UserDelete)
}
