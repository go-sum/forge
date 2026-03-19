package form

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// CheckboxProps configures a checkbox input.
type CheckboxProps struct {
	ID       string
	Name     string
	Value    string
	Checked  bool
	Disabled bool
	Extra    []g.Node
}

// Checkbox renders a styled <input type="checkbox">.
func Checkbox(p CheckboxProps) g.Node {
	nodes := []g.Node{
		h.Class("peer size-4 shrink-0 rounded-[4px] border border-input shadow-xs focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:border-ring disabled:cursor-not-allowed disabled:opacity-50 checked:bg-primary checked:text-primary-foreground checked:border-primary appearance-none cursor-pointer transition-shadow"),
		h.Type("checkbox"),
	}
	if p.ID != "" {
		nodes = append(nodes, h.ID(p.ID))
	}
	if p.Name != "" {
		nodes = append(nodes, h.Name(p.Name))
	}
	if p.Value != "" {
		nodes = append(nodes, h.Value(p.Value))
	}
	if p.Checked {
		nodes = append(nodes, h.Checked())
	}
	if p.Disabled {
		nodes = append(nodes, h.Disabled())
	}
	nodes = append(nodes, g.Group(p.Extra))
	return h.Input(nodes...)
}
