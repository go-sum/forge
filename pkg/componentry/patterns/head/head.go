// Package head provides reusable HTML head builders for page layouts.
// It stays generic by accepting concrete asset/meta values from callers rather
// than importing application-specific config or asset packages.
package head

import (
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// MetaProps configures the standard metadata emitted in <head>.
type MetaProps struct {
	Title       string
	Description string
	Keywords    []string
	FaviconHref string
}

// Stylesheet configures a single <link rel="stylesheet"> tag.
type Stylesheet struct {
	Href string
}

// Script configures a single external <script> tag.
type Script struct {
	Src   string
	Defer bool
	Async bool
}

// Props configures a complete <head> element.
type Props struct {
	Meta        MetaProps
	Stylesheets []Stylesheet
	Extra       []g.Node
	Scripts     []Script
}

// Head renders a complete <head> from typed metadata, assets, and caller extras.
func Head(p Props) g.Node {
	nodes := []g.Node{
		Metatags(p.Meta),
		CSS(p.Stylesheets...),
	}
	if len(p.Extra) > 0 {
		nodes = append(nodes, g.Group(p.Extra))
	}
	nodes = append(nodes, JS(p.Scripts...))
	return h.Head(g.Group(nodes))
}

// Metatags renders the standard metadata block for a document head.
func Metatags(p MetaProps) g.Node {
	nodes := []g.Node{
		h.Meta(h.Charset("utf-8")),
		h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
		h.TitleEl(g.Text(p.Title)),
	}
	if p.FaviconHref != "" {
		nodes = append(nodes, h.Link(h.Rel("icon"), h.Href(p.FaviconHref)))
	}
	if p.Description != "" {
		nodes = append(nodes, h.Meta(h.Name("description"), h.Content(p.Description)))
	}
	if len(p.Keywords) > 0 {
		nodes = append(nodes, h.Meta(h.Name("keywords"), h.Content(strings.Join(p.Keywords, ", "))))
	}
	return g.Group(nodes)
}

// CSS renders stylesheet links.
func CSS(stylesheets ...Stylesheet) g.Node {
	if len(stylesheets) == 0 {
		return g.Text("")
	}
	nodes := make([]g.Node, 0, len(stylesheets))
	for _, stylesheet := range stylesheets {
		if stylesheet.Href == "" {
			continue
		}
		nodes = append(nodes, h.Link(
			h.Rel("stylesheet"),
			h.Href(stylesheet.Href),
			h.Type("text/css"),
		))
	}
	if len(nodes) == 0 {
		return g.Text("")
	}
	return g.Group(nodes)
}

// JS renders external script tags.
func JS(scripts ...Script) g.Node {
	if len(scripts) == 0 {
		return g.Text("")
	}
	nodes := make([]g.Node, 0, len(scripts))
	for _, script := range scripts {
		if script.Src == "" {
			continue
		}
		attrs := []g.Node{h.Src(script.Src)}
		if script.Defer {
			attrs = append(attrs, h.Defer())
		}
		if script.Async {
			attrs = append(attrs, h.Async())
		}
		nodes = append(nodes, h.Script(attrs...))
	}
	if len(nodes) == 0 {
		return g.Text("")
	}
	return g.Group(nodes)
}

