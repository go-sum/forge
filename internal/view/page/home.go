package page

import (
	"starter/internal/view"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// HomePage renders the application landing page inside the base layout.
func HomePage(req view.Request) g.Node {
	return req.Page(
		"Home",
		h.Div(
			h.Class("flex flex-col items-center justify-center py-24 gap-4"),
			h.H1(h.Class("text-2xl font-bold"), g.Text("Welcome")),
			h.P(h.Class("text-muted-foreground"), g.Text("A Go starter with Echo, HTMX, and Tailwind.")),
		),
	)
}
