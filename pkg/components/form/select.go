package form

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Option is a single <option> value/label pair.
type Option struct {
	Value string
	Label string
}

// SelectProps configures a <select> element.
type SelectProps struct {
	ID       string
	Name     string
	Multiple bool
	Disabled bool
	HasError bool
	Options  []Option
	Selected string
	Extra    []g.Node
}

func selectClass(hasError bool) string {
	base := "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-base shadow-xs transition-colors outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"
	if hasError {
		base += " border-destructive ring-destructive/20"
	}
	return base
}

// Select renders a native <select> dropdown.
func Select(p SelectProps) g.Node {
	nodes := []g.Node{h.Class(selectClass(p.HasError))}
	if p.ID != "" {
		nodes = append(nodes, h.ID(p.ID))
	}
	if p.Name != "" {
		nodes = append(nodes, h.Name(p.Name))
	}
	if p.Disabled {
		nodes = append(nodes, h.Disabled())
	}
	if p.HasError {
		nodes = append(nodes, g.Attr("aria-invalid", "true"))
	}
	nodes = append(nodes, g.Group(p.Extra))
	for _, opt := range p.Options {
		optNodes := []g.Node{h.Value(opt.Value), g.Text(opt.Label)}
		if opt.Value == p.Selected {
			optNodes = append([]g.Node{h.Selected()}, optNodes...)
		}
		nodes = append(nodes, h.Option(optNodes...))
	}
	return h.Select(nodes...)
}
