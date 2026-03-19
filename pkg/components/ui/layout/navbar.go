// Package layout provides structural shell components for page layout.
package layout

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// NavbarProps configures a top navigation bar.
type NavbarProps struct {
	Brand      g.Node
	StartItems []g.Node
	EndItems   []g.Node
}

// Navbar renders a top navigation bar with start/center/end slots.
func Navbar(p NavbarProps) g.Node {
	return h.Div(
		h.Class("w-full border-b bg-background"),
		h.Div(
			h.Class("container mx-auto flex h-14 items-center px-4"),
			h.Div(h.Class("mr-4 flex"), p.Brand),
			h.Div(h.Class("flex flex-1 items-center gap-2"), g.Group(p.StartItems)),
			h.Div(h.Class("flex items-center gap-2 ml-auto"), g.Group(p.EndItems)),
		),
	)
}
