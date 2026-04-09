package page

import (
	"github.com/go-sum/componentry/ui/core"
	"github.com/go-sum/forge/internal/view"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// HomePage renders the application landing page inside the base layout.
func HomePage(req view.Request) g.Node {
	primary := core.Button(core.ButtonProps{
		Label:   "Browse Components",
		Href:    req.Path("components.list"),
		Variant: core.VariantOutline,
	})
	secondary := core.Button(core.ButtonProps{
		Label: "Sign In",
		Href:  req.Path("signin.get"),
	})
	if req.IsAuthenticated {
		secondary = core.Button(core.ButtonProps{
			Label: "Manage Users",
			Href:  req.Path("admin.user.list"),
		})
	}

	return req.Page(
		"Home",
		h.Div(
			h.Class("mx-auto flex max-w-3xl flex-col items-center justify-center gap-8 py-24 text-center"),
			h.Div(
				h.Class("space-y-4"),
				h.P(
					h.Class("text-sm font-medium uppercase tracking-[0.2em] text-muted-foreground"),
					g.Text("Modern Web Starter"),
				),
				h.H1(h.Class("text-2xl font-bold"), g.Text("Build server-rendered apps without giving up interaction quality.")),
				h.P(
					h.Class("mx-auto max-w-2xl text-sm text-muted-foreground"),
					g.Text("This starter combines Echo, Gomponents, HTMX, and reusable UI packages so pages stay fast, maintainable, and visually consistent."),
				),
			),
			h.Div(
				h.Class("flex flex-col gap-3 sm:flex-row"),
				secondary,
				primary,
			),
		),
	)
}
