package handler

import (
	"github.com/go-sum/componentry/examples"
	"github.com/go-sum/forge/internal/view"

	"github.com/labstack/echo/v5"
)

// ComponentExamples renders the component library reference page.
func (h *Handler) ComponentExamples(c *echo.Context) error {
	req := h.request(c)
	return view.Render(c, req, req.Page("Component Examples", examples.Page()), nil)
}
