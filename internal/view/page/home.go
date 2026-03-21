package page

import (
	"starter/internal/view/layout"
	"starter/pkg/components/patterns/flash"
	uilayout "starter/pkg/components/ui/layout"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// HomeProps configures the landing page.
type HomeProps struct {
	CSRFToken       string
	IsAuthenticated bool
	Flash           []flash.Message
	NavConfig       uilayout.NavConfig
}

// HomePage renders the application landing page inside the base layout.
func HomePage(p HomeProps) g.Node {
	return layout.Page(layout.Props{
		Title:           "Home",
		CSRFToken:       p.CSRFToken,
		IsAuthenticated: p.IsAuthenticated,
		Flash:           p.Flash,
		NavConfig:       p.NavConfig,
		Children: []g.Node{
			h.Div(
				h.Class("flex flex-col items-center justify-center py-24 gap-4"),
				h.H1(h.Class("text-2xl font-bold"), g.Text("Welcome")),
				h.P(h.Class("text-muted-foreground"), g.Text("A Go starter with Echo, HTMX, and Tailwind.")),
			),
		},
	})
}
