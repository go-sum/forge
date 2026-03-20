package handler

import (
	custommw "starter/internal/middleware"
	"starter/pkg/auth"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes maps all application routes onto e.
// Public file serving and dynamic routes are declared here so the full URL
// surface is visible in one place.
func (h *Handler) RegisterRoutes(e *echo.Echo, sessions *auth.SessionManager, publicPrefix, publicDir string) {
	e.Static(publicPrefix, publicDir)

	e.GET("/health", h.HealthCheck)
	e.GET("/_components", h.ComponentExamples)
	e.GET("/", h.Home)

	e.GET("/login", h.LoginPage)
	e.POST("/login", h.Login)
	e.GET("/register", h.RegisterPage)
	e.POST("/register", h.Register)

	protected := e.Group("")
	protected.Use(custommw.RequireAuth(sessions))
	protected.POST("/logout", h.Logout)
	protected.GET("/users", h.UserList)
	protected.GET("/users/:id/edit", h.UserEditForm)
	protected.GET("/users/:id/row", h.UserRow)
	protected.PUT("/users/:id", h.UserUpdate)
	protected.DELETE("/users/:id", h.UserDelete)
}
