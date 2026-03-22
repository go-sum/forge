package form

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// FileUploadProps configures a styled drag-and-drop file input zone.
type FileUploadProps struct {
	ID       string
	Name     string
	Accept   string // e.g. "image/*,.pdf" — passed to accept attribute; blank → omitted
	Multiple bool
	Disabled bool
	HasError bool
	Prompt   string   // label text inside drop zone; defaults to "Drag & drop or click to browse"
	Extra    []g.Node // extra attributes on the <label> wrapper
}

func fileUploadClass(hasError bool) string {
	base := "flex flex-col items-center justify-center gap-3 rounded-md border-2 border-dashed " +
		"border-input bg-transparent p-8 text-center transition-colors cursor-pointer " +
		"focus-within:border-ring focus-within:ring-ring/50 focus-within:ring-[3px] " +
		"data-[dragging]:border-primary data-[dragging]:bg-primary/5 " +
		"has-[:disabled]:cursor-not-allowed has-[:disabled]:opacity-50"
	if hasError {
		base += " border-destructive data-[dragging]:border-destructive"
	}
	return base
}

// FileUpload renders a styled drag-and-drop drop zone wrapping a hidden <input type="file">.
// static/js/components.js delegates drag/drop/change events via [data-file-upload].
func FileUpload(p FileUploadProps, children ...g.Node) g.Node {
	// Label (drop zone) attributes
	nodes := []g.Node{
		h.Class(fileUploadClass(p.HasError)),
		g.Attr("data-file-upload", ""),
	}
	if p.ID != "" {
		nodes = append(nodes, h.For(p.ID))
	}
	nodes = append(nodes, g.Group(p.Extra))

	// Hidden file input
	inputNodes := []g.Node{
		h.Type("file"),
		h.Class("sr-only"),
	}
	if p.ID != "" {
		inputNodes = append(inputNodes, h.ID(p.ID))
	}
	if p.Name != "" {
		inputNodes = append(inputNodes, h.Name(p.Name))
	}
	if p.Accept != "" {
		inputNodes = append(inputNodes, g.Attr("accept", p.Accept))
	}
	if p.Multiple {
		inputNodes = append(inputNodes, g.Attr("multiple", ""))
	}
	if p.Disabled {
		inputNodes = append(inputNodes, h.Disabled())
	}
	if p.HasError {
		inputNodes = append(inputNodes, g.Attr("aria-invalid", "true"))
	}
	nodes = append(nodes, h.Input(inputNodes...))

	// Prompt: children override the default text
	prompt := p.Prompt
	if prompt == "" {
		prompt = "Drag & drop or click to browse"
	}
	if len(children) > 0 {
		nodes = append(nodes, g.Group(children))
	} else {
		nodes = append(nodes, h.Span(
			h.Class("text-sm font-medium pointer-events-none"),
			g.Text(prompt),
		))
	}

	// Filename display (updated by JS)
	nodes = append(nodes, h.Span(
		g.Attr("data-file-name", ""),
		g.Attr("aria-live", "polite"),
		h.Class("text-xs text-muted-foreground pointer-events-none"),
	))

	return h.Label(nodes...)
}
