// Package email provides Gomponents-based building blocks for composing HTML emails.
// All components use table-based layout and inline styles for maximum email client
// compatibility. Use [Layout] as the outer shell and [P], [H1] for content elements.
package email

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Layout wraps body in an email-safe HTML shell using table-based layout.
// The title appears in the document <title> element.
func Layout(title string, body g.Node) g.Node {
	return g.Group([]g.Node{
		g.Raw(`<!DOCTYPE html>`),
		h.HTML(
			h.Lang("en"),
			h.Head(
				h.Meta(g.Attr("charset", "utf-8")),
				h.Meta(
					g.Attr("name", "viewport"),
					g.Attr("content", "width=device-width,initial-scale=1"),
				),
				h.TitleEl(g.Text(title)),
			),
			h.Body(
				g.Attr("style", "margin:0;padding:0;background-color:#f4f4f4;font-family:Arial,Helvetica,sans-serif;"),
				h.Table(
					g.Attr("width", "100%"),
					g.Attr("cellpadding", "0"),
					g.Attr("cellspacing", "0"),
					g.Attr("role", "presentation"),
					g.Attr("style", "background-color:#f4f4f4;"),
					h.TBody(
						h.Tr(
							h.Td(
								g.Attr("align", "center"),
								g.Attr("style", "padding:40px 16px;"),
								h.Table(
									g.Attr("width", "600"),
									g.Attr("cellpadding", "0"),
									g.Attr("cellspacing", "0"),
									g.Attr("role", "presentation"),
									g.Attr("style", "background-color:#ffffff;border-radius:4px;max-width:600px;width:100%;"),
									h.TBody(
										h.Tr(
											h.Td(
												g.Attr("style", "padding:40px;color:#1a1a1a;font-size:14px;line-height:1.6;"),
												body,
											),
										),
									),
								),
							),
						),
					),
				),
			),
		),
	})
}

// H1 renders a heading with email-safe inline styling.
func H1(text string) g.Node {
	return h.H1(
		g.Attr("style", "margin:0 0 24px;font-size:24px;font-weight:bold;color:#1a1a1a;"),
		g.Text(text),
	)
}

// P renders a paragraph with email-safe inline styling.
func P(text string) g.Node {
	return h.P(
		g.Attr("style", "margin:0 0 16px;"),
		g.Text(text),
	)
}
