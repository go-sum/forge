package handler

import (
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/page"

	"github.com/labstack/echo/v5"
)

// Home renders the application landing page.
func (h *Handler) Home(c *echo.Context) error {
	req := h.request(c)
	return view.Render(c, req, page.HomePage(req), nil)
}
