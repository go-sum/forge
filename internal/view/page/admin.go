package page

import (
	"github.com/go-sum/componentry/ui/core"
	"github.com/go-sum/forge/internal/view"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// AdminElevatePage renders the admin elevation form, shown only when no admin
// exists yet.
func AdminElevatePage(req view.Request) g.Node {
	return req.Page(
		"Become Admin",
		h.Div(
			h.Class("mx-auto max-w-2xl py-12"),
			h.Div(
				h.Class("space-y-2 mb-8"),
				h.H1(h.Class("text-2xl font-bold"), g.Text("Become Admin")),
				h.P(
					h.Class("text-sm text-muted-foreground"),
					g.Text("No admin account exists yet. Elevate your account to take ownership of this application."),
				),
			),
			h.Form(
				h.Method("post"),
				h.Action(req.Path("account.admin.post")),
				h.Class("space-y-4"),
				h.Input(h.Type("hidden"), h.Name(req.CSRFFieldName), h.Value(req.CSRFToken)),
				h.Div(
					h.Class("flex justify-end"),
					core.Button(core.ButtonProps{
						Label: "Elevate to Admin",
						Type:  "submit",
					}),
				),
			),
		),
	)
}
