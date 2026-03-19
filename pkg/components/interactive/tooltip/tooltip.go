// Package tooltip provides a CSS-only tooltip using Tailwind's group/group-hover pattern.
// No JavaScript required — visibility is controlled entirely by CSS.
package tooltip

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Root renders the tooltip root container. The `group` class propagates
// hover state to descendant elements via Tailwind's group-hover: variant.
func Root(children ...g.Node) g.Node {
	return h.Div(
		h.Class("relative inline-flex group"),
		g.Group(children),
	)
}

// Trigger renders a pass-through wrapper around the trigger element.
func Trigger(children ...g.Node) g.Node {
	return h.Div(
		h.Class("contents"),
		g.Group(children),
	)
}

// Content renders the tooltip panel. It is hidden by default and shown
// on hover via the group-hover:block Tailwind utility on the Root.
func Content(children ...g.Node) g.Node {
	return h.Div(
		h.Class("absolute bottom-full mb-2 left-1/2 -translate-x-1/2 z-50 rounded-md border bg-popover px-3 py-1.5 text-xs text-popover-foreground shadow-md pointer-events-none whitespace-nowrap hidden group-hover:block"),
		g.Group(children),
	)
}
