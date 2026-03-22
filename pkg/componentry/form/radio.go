package form

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Radio renders a styled radio button as a composite: a hidden peer <input> plus
// visual ring and dot spans driven by peer-checked CSS. CSS pseudo-elements
// cannot apply to <input> directly, so the dot is a real sibling span.
func Radio(p RadioProps) g.Node {
	return h.Span(
		h.Class("relative inline-flex size-4 shrink-0 cursor-pointer"),
		h.Input(buildToggleInput("radio", p)...),
		// Ring — outer circle border, colour driven by peer state.
		h.Span(h.Class("absolute inset-0 rounded-full border border-input bg-transparent transition-colors peer-checked:border-primary peer-focus-visible:ring-[3px] peer-focus-visible:ring-ring/50 peer-disabled:opacity-50")),
		// Dot — inner filled circle, visible only when checked.
		h.Span(h.Class("absolute inset-0 m-auto size-2 rounded-full bg-transparent transition-colors peer-checked:bg-primary")),
	)
}
