package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// HealthCheck reports database reachability.
// Returns 200 when the DB is reachable, 503 otherwise.
func (h *Handler) HealthCheck(c *echo.Context) error {
	status, code := "ok", http.StatusOK
	if err := h.checkHealth(c.Request().Context()); err != nil {
		status, code = "error", http.StatusServiceUnavailable
	}
	return c.JSON(code, map[string]string{"status": status})
}
