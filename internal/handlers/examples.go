package handlers

import (
	"starter/internal/view/layout"
	"starter/pkg/components/examples"
	"starter/pkg/render"

	g "maragu.dev/gomponents"

	"github.com/labstack/echo/v5"
)

// ComponentExamples renders the component library reference page.
func (h *Handler) ComponentExamples(c *echo.Context) error {
	return render.Component(c, layout.Page(layout.Props{
		Title:     "Component Examples",
		CSRFToken: h.csrfToken(c),
		Children:  []g.Node{examples.Page()},
	}))
}
