package layout

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Footer renders the site-wide page footer with a copyright notice and app version.
func Footer(p Props) g.Node {
	return h.Footer(
		h.Class("border-t border-border bg-background"),
		h.Div(
			h.Class("container mx-auto px-4 py-6 text-center text-sm text-muted-foreground"),
			g.Textf("© %d %s. All rights reserved.", p.CopyrightYear, p.NavConfig.Brand.Label),
			g.If(p.AppVersion != "", h.Span(
				h.Class("ml-2 opacity-50 italic text-xs"),
				g.Text(p.AppVersion),
			)),
		),
	)
}
