package form

import (
	"strconv"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// TextareaProps configures a multi-line text field.
type TextareaProps struct {
	ID          string
	Name        string
	Placeholder string
	Value       string
	Rows        int
	Disabled    bool
	Readonly    bool
	HasError    bool
	Extra       []g.Node
}

func textareaClass(hasError bool) string {
	base := "flex min-h-[60px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-base shadow-xs transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"
	if hasError {
		base += " border-destructive ring-destructive/20"
	}
	return base
}

// Textarea renders a multi-line <textarea>.
func Textarea(p TextareaProps) g.Node {
	nodes := []g.Node{h.Class(textareaClass(p.HasError))}
	if p.ID != "" {
		nodes = append(nodes, h.ID(p.ID))
	}
	if p.Name != "" {
		nodes = append(nodes, h.Name(p.Name))
	}
	if p.Placeholder != "" {
		nodes = append(nodes, h.Placeholder(p.Placeholder))
	}
	if p.Rows > 0 {
		nodes = append(nodes, h.Rows(strconv.Itoa(p.Rows)))
	}
	if p.Disabled {
		nodes = append(nodes, h.Disabled())
	}
	if p.Readonly {
		nodes = append(nodes, h.ReadOnly())
	}
	if p.HasError {
		nodes = append(nodes, g.Attr("aria-invalid", "true"))
	}
	nodes = append(nodes, g.Group(p.Extra))
	nodes = append(nodes, g.Text(p.Value))
	return h.Textarea(nodes...)
}
