package form

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Checkbox renders a styled checkbox as a composite: a hidden peer <input> plus
// visual box and checkmark spans driven by peer-checked CSS. CSS pseudo-elements
// cannot apply to <input> directly, so the checkmark is a real SVG sibling.
func Checkbox(p CheckboxProps) g.Node {
	return h.Span(
		h.Class("relative inline-flex size-4 shrink-0 cursor-pointer"),
		h.Input(buildToggleInput("checkbox", p)...),
		// Box — border becomes filled when checked.
		h.Span(h.Class("absolute inset-0 rounded-[4px] border border-input bg-transparent transition-colors peer-checked:border-primary peer-checked:bg-primary peer-focus-visible:ring-[3px] peer-focus-visible:ring-ring/50 peer-disabled:opacity-50")),
		// Checkmark — inline SVG path, visible only when checked.
		h.SVG(
			h.Class("absolute inset-0 m-auto size-3 hidden peer-checked:block text-primary-foreground"),
			g.Attr("viewBox", "0 0 10 10"),
			g.Attr("fill", "none"),
			g.Attr("stroke", "currentColor"),
			g.Attr("stroke-width", "1.5"),
			g.Attr("aria-hidden", "true"),
			g.El("path", g.Attr("d", "M2 6l3 3 5-5")),
		),
	)
}
