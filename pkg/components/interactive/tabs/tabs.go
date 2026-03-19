// Package tabs provides an accessible tabs component using ARIA roles and data attributes.
// The initial active state is rendered server-side (SSR-ready, works without JS).
// app.js initTabs() attaches click handlers for switching panels at runtime.
package tabs

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Root renders the root tabs container. defaultTab sets the initially active tab value.
// The value is stored in data-tabs so initTabs() in app.js can read it.
func Root(defaultTab string, children ...g.Node) g.Node {
	return h.Div(
		g.Attr("data-tabs", defaultTab),
		h.Class("w-full"),
		g.Group(children),
	)
}

// List renders the tab button bar.
func List(children ...g.Node) g.Node {
	return h.Div(
		h.Role("tablist"),
		h.Class("inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground"),
		g.Group(children),
	)
}

// Trigger renders a single tab button. isDefault marks the initially active tab,
// which is reflected via aria-selected and active styling without requiring JS.
func Trigger(value string, isDefault bool, children ...g.Node) g.Node {
	cls := "inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1 text-sm font-medium transition-all focus-visible:outline-none focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-50"
	ariaSelected := "false"
	if isDefault {
		ariaSelected = "true"
	}
	nodes := []g.Node{
		h.Type("button"),
		h.Role("tab"),
		g.Attr("data-tab", value),
		g.Attr("aria-selected", ariaSelected),
		h.Class(cls),
	}
	if isDefault {
		nodes = append(nodes, h.Class("bg-background text-foreground shadow"))
	}
	nodes = append(nodes, g.Group(children))
	return h.Button(nodes...)
}

// Content renders the panel for a specific tab value. isDefault controls
// whether the panel is visible on initial render (hidden attr set when false).
func Content(value string, isDefault bool, children ...g.Node) g.Node {
	nodes := []g.Node{
		h.Role("tabpanel"),
		g.Attr("data-tab", value),
		h.Class("mt-2"),
	}
	if !isDefault {
		nodes = append(nodes, g.Attr("hidden", ""))
	}
	nodes = append(nodes, g.Group(children))
	return h.Div(nodes...)
}
