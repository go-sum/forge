package errorpage

import (
	"fmt"
	"net/http"

	"starter/internal/view"
	"starter/pkg/components/ui/core"
	"starter/pkg/components/ui/data"
	"starter/pkg/components/ui/feedback"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Props configures the full-page application error view.
type Props struct {
	Status          int
	Title           string
	Message         string
	RequestID       string
	Debug           bool
	TechnicalDetail string
	HomePath        string
}

// Page renders a user-facing error page with optional development details.
func Page(req view.Request, p Props) g.Node {
	title := p.Title
	if title == "" {
		title = http.StatusText(p.Status)
	}
	homePath := p.HomePath
	if homePath == "" {
		homePath = "/"
	}

	return req.Page(
		title,
		h.Div(
			h.Class("mx-auto flex max-w-2xl flex-col gap-6 py-16"),
			data.Card.Root(
				data.Card.Header(
					h.Div(
						h.Class("flex items-start justify-between gap-4"),
						h.Div(
							h.Class("space-y-1"),
							data.Card.Title(g.Text(title)),
							data.Card.Description(g.Textf("%d %s", p.Status, title)),
						),
						core.Badge(core.BadgeProps{
							Variant:  core.BadgeSecondary,
							Children: []g.Node{g.Text(fmt.Sprintf("HTTP %d", p.Status))},
						}),
					),
				),
				data.Card.Content(
					h.Div(
						h.Class("space-y-4"),
						feedback.Alert.Root(
							feedback.AlertProps{Variant: alertVariantForStatus(p.Status)},
							feedback.Alert.Description(g.Text(p.Message)),
						),
						h.Div(
							h.Class("flex flex-wrap gap-3"),
							core.Button(core.ButtonProps{
								Label:   "Return Home",
								Href:    homePath,
								Variant: core.VariantDefault,
							}),
						),
						g.If(
							p.RequestID != "",
							h.P(
								h.Class("text-sm text-muted-foreground"),
								g.Text("Request ID: "),
								h.Code(h.Class("font-medium text-foreground"), g.Text(p.RequestID)),
							),
						),
						g.If(p.Debug && p.TechnicalDetail != "", debugDetail(p.TechnicalDetail)),
					),
				),
			),
		),
	)
}

func alertVariantForStatus(status int) feedback.AlertVariant {
	if status >= http.StatusInternalServerError {
		return feedback.AlertDestructive
	}
	return feedback.AlertDefault
}

func debugDetail(detail string) g.Node {
	return h.Details(
		h.Class("rounded-lg border bg-card p-4 text-sm"),
		h.Summary(h.Class("cursor-pointer font-medium"), g.Text("Technical Detail")),
		h.Pre(
			h.Class("mt-3 overflow-x-auto whitespace-pre-wrap break-words text-muted-foreground"),
			g.Text(detail),
		),
	)
}
