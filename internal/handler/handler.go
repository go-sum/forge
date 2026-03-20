// Package handler contains the HTTP transport layer. Each handler method
// parses request data, delegates to a service, and renders a response.
package handler

import (
	"starter/internal/service"
	"starter/pkg/auth"
	"starter/pkg/validate"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// Handler holds the transport layer's dependencies.
type Handler struct {
	services      *service.Services
	sessions      *auth.SessionManager
	validator     *validate.Validator
	pool          *pgxpool.Pool
	csrfFieldName string
}

// New constructs a Handler with all required dependencies.
func New(
	services *service.Services,
	sessions *auth.SessionManager,
	validator *validate.Validator,
	pool *pgxpool.Pool,
	csrfFieldName string,
) *Handler {
	return &Handler{
		services:      services,
		sessions:      sessions,
		validator:     validator,
		pool:          pool,
		csrfFieldName: csrfFieldName,
	}
}

// csrfToken reads the CSRF token stored in context by Echo's CSRF middleware.
// Uses DefaultCSRFConfig.ContextKey to avoid hardcoding the "csrf" string.
func (h *Handler) csrfToken(c *echo.Context) string {
	v, _ := c.Get(middleware.DefaultCSRFConfig.ContextKey).(string)
	return v
}
