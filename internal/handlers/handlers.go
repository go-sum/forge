// Package handlers contains the HTTP transport layer. Each handler method
// parses request data, delegates to a service, and renders a response.
// Methods that depend on unbuilt layers (auth, views) are stubbed and return
// echo.ErrNotImplemented until those phases are complete.
package handlers

import (
	"starter/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// Handler holds the transport layer's dependencies. The struct grows as domain
// layers are built (T0302 auth, T0901–T0903 services).
type Handler struct {
	pool          *pgxpool.Pool
	config        *config.Config
	csrfFieldName string
}

func New(pool *pgxpool.Pool, cfg *config.Config) *Handler {
	return &Handler{
		pool:          pool,
		config:        cfg,
		csrfFieldName: cfg.Server.CSRFCookieName,
	}
}

// csrfToken reads the CSRF token stored in context by Echo's CSRF middleware.
// Uses DefaultCSRFConfig.ContextKey to avoid hardcoding the "csrf" string.
func (h *Handler) csrfToken(c *echo.Context) string {
	v, _ := c.Get(middleware.DefaultCSRFConfig.ContextKey).(string)
	return v
}
