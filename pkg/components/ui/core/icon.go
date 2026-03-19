package core

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// IconProps configures an Icon rendered as an SVG <use> reference into a sprite file.
type IconProps struct {
	Src   string   // sprite file path, e.g. "/static/img/svg/lucide-icons.svg"
	ID    string   // symbol id, e.g. "chevron-down"
	Size  string   // Tailwind size class; defaults to "size-4"
	Label string   // aria-label text; empty → aria-hidden="true" (decorative)
	Extra []g.Node // additional attributes on the outer <svg> element
}

// Icon renders an accessible <svg><use href="sprite#id"/></svg> element.
// Decorative icons (no Label) get aria-hidden="true"; labelled icons get role="img".
func Icon(p IconProps) g.Node {
	size := p.Size
	if size == "" {
		size = "size-4"
	}

	nodes := []g.Node{h.Class(size)}
	if p.Label != "" {
		nodes = append(nodes, g.Attr("role", "img"), g.Attr("aria-label", p.Label))
	} else {
		nodes = append(nodes, g.Attr("aria-hidden", "true"))
	}
	nodes = append(nodes, g.Group(p.Extra))
	nodes = append(nodes, g.El("use", h.Href(p.Src+"#"+p.ID)))

	return h.SVG(nodes...)
}
