package form

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// ToggleProps is the shared props type for Checkbox and Radio.
type ToggleProps struct {
	ID       string
	Name     string
	Value    string
	Checked  bool
	Disabled bool
	Required bool
	Extra    []g.Node
}

// CheckboxProps and RadioProps are aliases so existing callers compile unchanged.
type CheckboxProps = ToggleProps
type RadioProps = ToggleProps

// buildToggleInput builds the sr-only peer <input> nodes shared by Checkbox and Radio.
func buildToggleInput(inputType string, p ToggleProps) []g.Node {
	nodes := []g.Node{
		h.Class("sr-only peer"),
		h.Type(inputType),
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
	if p.Required {
		nodes = append(nodes, g.Attr("required", ""))
	}
	nodes = append(nodes, g.Group(p.Extra))
	return nodes
}
