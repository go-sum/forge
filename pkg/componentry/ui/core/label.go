package core

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// LabelProps configures a form label.
type LabelProps struct {
	For   string
	Error string
	Extra []g.Node
}

// Label renders a <label> element. The Error field adds a destructive colour.
func Label(p LabelProps, children ...g.Node) g.Node {
	cls := "text-sm font-medium leading-none inline-block"
	if p.Error != "" {
		cls += " text-destructive"
	}
	nodes := []g.Node{h.Class(cls)}
	if p.For != "" {
		nodes = append(nodes, h.For(p.For))
	}
	nodes = append(nodes, g.Group(p.Extra))
	nodes = append(nodes, g.Group(children))
	return h.Label(nodes...)
}
