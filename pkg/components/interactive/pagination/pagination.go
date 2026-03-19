// Package pagination provides navigation pagination components.
package pagination

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	componenticons "starter/pkg/components/icons"
	iconrender "starter/pkg/components/icons/render"
	core "starter/pkg/components/ui/core"
)

// Root renders a <nav role="navigation"> wrapper.
func Root(children ...g.Node) g.Node {
	return h.Nav(
		h.Role("navigation"),
		g.Attr("aria-label", "pagination"),
		h.Class("flex flex-wrap justify-center"),
		g.Group(children),
	)
}

// Content renders the <ul> flex row.
func Content(children ...g.Node) g.Node {
	return h.Ul(
		h.Class("flex flex-row items-center gap-1"),
		g.Group(children),
	)
}

// Item renders a <li> wrapper.
func Item(children ...g.Node) g.Node {
	return h.Li(g.Group(children))
}

func paginationLinkBase(isActive bool) string {
	base := "inline-flex items-center justify-center size-9 rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
	if isActive {
		return base + " border border-border bg-background shadow-xs"
	}
	return base + " hover:bg-accent hover:text-accent-foreground"
}

// Link renders a page number link.
func Link(href string, isActive bool, children ...g.Node) g.Node {
	nodes := []g.Node{
		h.Class(paginationLinkBase(isActive)),
		h.Href(href),
	}
	if isActive {
		nodes = append(nodes, g.Attr("aria-current", "page"))
	}
	nodes = append(nodes, g.Group(children))
	return h.A(nodes...)
}

// Previous renders a "previous" navigation button.
func Previous(href string, disabled bool) g.Node {
	cls := "inline-flex items-center gap-1 px-2.5 h-9 rounded-md text-sm font-medium transition-colors"
	if disabled {
		cls += " pointer-events-none opacity-50"
	} else {
		cls += " hover:bg-accent hover:text-accent-foreground"
	}
	nodes := []g.Node{h.Class(cls)}
	if !disabled {
		nodes = append(nodes, h.Href(href))
	}
	nodes = append(nodes, g.Attr("aria-label", "Go to previous page"))
	nodes = append(nodes,
		core.Icon(iconrender.PropsFor(componenticons.ChevronLeft, core.IconProps{})),
		h.Span(g.Text("Previous")),
	)
	return h.A(nodes...)
}

// Next renders a "next" navigation button.
func Next(href string, disabled bool) g.Node {
	cls := "inline-flex items-center gap-1 px-2.5 h-9 rounded-md text-sm font-medium transition-colors"
	if disabled {
		cls += " pointer-events-none opacity-50"
	} else {
		cls += " hover:bg-accent hover:text-accent-foreground"
	}
	nodes := []g.Node{h.Class(cls)}
	if !disabled {
		nodes = append(nodes, h.Href(href))
	}
	nodes = append(nodes, g.Attr("aria-label", "Go to next page"))
	nodes = append(nodes,
		h.Span(g.Text("Next")),
		core.Icon(iconrender.PropsFor(componenticons.ChevronRight, core.IconProps{})),
	)
	return h.A(nodes...)
}

// Ellipsis renders a "…" placeholder for skipped page ranges.
func Ellipsis() g.Node {
	return h.Span(
		h.Class("flex size-9 items-center justify-center"),
		g.Attr("aria-hidden", "true"),
		g.Text("…"),
	)
}
