package handler

import (
	"net/http"

	"starter/internal/view/layout"
	"starter/internal/view/page"
	"starter/pkg/components/examples"
	"starter/pkg/components/patterns/flash"
	"starter/pkg/ctxkeys"
	"starter/pkg/render"

	g "maragu.dev/gomponents"

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
	userID, _ := c.Get(string(ctxkeys.UserID)).(string)
	flashMsgs, _ := flash.GetAll(c.Request(), c.Response())
	return render.Component(c, page.HomePage(page.HomeProps{
		CSRFToken:       h.csrfToken(c),
		IsAuthenticated: userID != "",
		Flash:           flashMsgs,
		NavConfig:       h.navConfig,
	}))
}

// ComponentExamples renders the component library reference page.
func (h *Handler) ComponentExamples(c *echo.Context) error {
	return render.Component(c, layout.Page(layout.Props{
		Title:     "Component Examples",
		CSRFToken: h.csrfToken(c),
		Children:  []g.Node{examples.Page()},
	}))
}
