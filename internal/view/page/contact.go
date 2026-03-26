package page

import (
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/partial/contactpartial"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// ContactPage renders the full-page contact-us view.
func ContactPage(req view.Request, data model.ContactFormData) g.Node {
	return req.Page(
		"Contact",
		h.Div(
			h.Class("mx-auto max-w-2xl py-12"),
			h.Div(
				h.Class("space-y-2 mb-8"),
				h.H1(h.Class("text-2xl font-bold"), g.Text("Contact us")),
				h.P(
					h.Class("text-sm text-muted-foreground"),
					g.Text("Fill out the form and we'll get back to you as soon as possible."),
				),
			),
			contactpartial.ContactForm(req, data),
		),
	)
}
