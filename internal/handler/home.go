package handler

import (
	"net/http"

	"starter/internal/view"
	"starter/internal/view/page"
	"starter/pkg/components/examples"

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

// Home renders the application landing page.
func (h *Handler) Home(c *echo.Context) error {
	req := h.request(c)
	return view.Render(c, req, page.HomePage(req), nil)
}

// ComponentExamples renders the component library reference page.
func (h *Handler) ComponentExamples(c *echo.Context) error {
	req := h.request(c)
	return view.Render(c, req, req.Page("Component Examples", examples.Page()), nil)
}
