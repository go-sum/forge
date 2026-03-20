package handler

import (
	"net/http"

	"starter/internal/view/layout"
	"starter/pkg/components/examples"
	"starter/pkg/components/patterns/flash"
	"starter/pkg/ctxkeys"
	"starter/pkg/database"
	"starter/pkg/render"

	g "maragu.dev/gomponents"
	"maragu.dev/gomponents/html"

	"github.com/labstack/echo/v5"
)

// HealthCheck reports database reachability.
// Returns 200 when the DB is reachable, 503 otherwise.
func (h *Handler) HealthCheck(c *echo.Context) error {
	status, code := "ok", http.StatusOK
	if err := database.CheckHealth(c.Request().Context(), h.pool); err != nil {
		status, code = "error", http.StatusServiceUnavailable
	}
	return c.JSON(code, map[string]string{"status": status})
}

// Home renders the application landing page.
func (h *Handler) Home(c *echo.Context) error {
	userID, _ := c.Get(string(ctxkeys.UserID)).(string)
	flashMsgs, _ := flash.GetAll(c.Request(), c.Response())
	return render.Component(c, layout.Page(layout.Props{
		Title:           "Home",
		CSRFToken:       h.csrfToken(c),
		IsAuthenticated: userID != "",
		Flash:           flashMsgs,
		Children: []g.Node{
			html.Div(
				html.Class("flex flex-col items-center justify-center py-24 gap-4"),
				html.H1(html.Class("text-4xl font-bold"), g.Text("Welcome")),
				html.P(html.Class("text-muted-foreground"), g.Text("A Go starter with Echo, HTMX, and Tailwind.")),
			),
		},
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
