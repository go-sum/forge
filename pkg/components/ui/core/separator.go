package core

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Orientation controls the separator's axis.
type Orientation string

const (
	OrientationHorizontal Orientation = "horizontal"
	OrientationVertical   Orientation = "vertical"
)

// Decoration controls the separator's line style.
type Decoration string

const (
	DecorationDefault Decoration = ""
	DecorationDashed  Decoration = "dashed"
	DecorationDotted  Decoration = "dotted"
)

// SeparatorProps configures a visual divider.
type SeparatorProps struct {
	Orientation Orientation
	Decoration  Decoration
	Label       string
	Extra       []g.Node
}

func decorationClass(d Decoration) string {
	switch d {
	case DecorationDashed:
		return " border-dashed"
	case DecorationDotted:
		return " border-dotted"
	default:
		return ""
	}
}

// Separator renders a horizontal or vertical divider with an optional centred label.
func Separator(p SeparatorProps) g.Node {
	if p.Orientation == OrientationVertical {
		return h.Div(
			h.Role("separator"),
			g.Attr("aria-orientation", "vertical"),
			h.Class("shrink-0 h-full"),
			g.Group(p.Extra),
			h.Div(
				h.Class("relative flex flex-col items-center h-full"),
				h.Span(
					h.Class("absolute h-full border-l w-[1px]"+decorationClass(p.Decoration)),
					g.Attr("aria-hidden", "true"),
				),
				g.If(p.Label != "",
					h.Span(h.Class("relative my-auto bg-background py-2 text-xs text-muted-foreground"), g.Text(p.Label)),
				),
			),
		)
	}
	return h.Div(
		h.Role("separator"),
		g.Attr("aria-orientation", "horizontal"),
		h.Class("shrink-0 w-full"),
		g.Group(p.Extra),
		h.Div(
			h.Class("relative flex items-center w-full"),
			h.Span(
				h.Class("absolute w-full border-t h-[1px]"+decorationClass(p.Decoration)),
				g.Attr("aria-hidden", "true"),
			),
			g.If(p.Label != "",
				h.Span(h.Class("relative mx-auto bg-background px-2 text-xs text-muted-foreground"), g.Text(p.Label)),
			),
		),
	)
}
