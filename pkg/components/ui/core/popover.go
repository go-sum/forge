package core

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type popoverNS struct{}

// Popover groups the CSS-first popover sub-components under a namespace.
// It uses native <details>/<summary> for click-toggle with no JavaScript.
// data-popover hooks the shared outside-click delegated listener in js/app.js.
var Popover popoverNS

// PopoverRootProps configures the popover root element.
type PopoverRootProps struct {
	ID string
	// Class overrides the default root positioning class ("relative inline-block").
	// Pass "relative inline-flex" for inline-flex contexts such as the tooltip click variant.
	Class string
	Extra []g.Node
}

// PopoverTriggerProps configures the <summary> trigger element.
type PopoverTriggerProps struct {
	// Class is appended to the base "list-none cursor-pointer" classes.
	// Apply button-like styling here instead of nesting a <button> inside <summary>.
	Class string
	Extra []g.Node
}

// PopoverContentProps configures the floating panel.
type PopoverContentProps struct {
	Width string // Tailwind width class, e.g. "w-64", "w-80"; default "w-72"
	Align string // "left" | "right" | "center" — panel alignment; default "left"
	Extra []g.Node
}

// Root renders <details data-popover class="relative inline-block [Class]">.
func (popoverNS) Root(p PopoverRootProps, children ...g.Node) g.Node {
	cls := p.Class
	if cls == "" {
		cls = "relative inline-block"
	}
	nodes := []g.Node{
		g.Attr("data-popover", ""),
		h.Class(cls),
	}
	if p.ID != "" {
		nodes = append(nodes, h.ID(p.ID))
	}
	nodes = append(nodes, g.Group(p.Extra))
	nodes = append(nodes, g.Group(children))
	return h.Details(nodes...)
}

// Trigger renders <summary class="list-none cursor-pointer [Class]">.
// The <summary> is the sole interactive element — do NOT nest <button> or <a>
// inside; doing so is invalid HTML and breaks the native click-to-toggle behaviour.
func (popoverNS) Trigger(p PopoverTriggerProps, children ...g.Node) g.Node {
	cls := "list-none cursor-pointer"
	if p.Class != "" {
		cls += " " + p.Class
	}
	nodes := []g.Node{
		h.Class(cls),
		g.Group(p.Extra),
		g.Group(children),
	}
	return h.Summary(nodes...)
}

// Content renders the positioned floating panel, shown when <details> is open.
// The browser UA stylesheet hides non-<summary> children of a closed <details>
// element natively — no hidden/group-open Tailwind classes are required.
func (popoverNS) Content(p PopoverContentProps, children ...g.Node) g.Node {
	width := p.Width
	if width == "" {
		width = "w-72"
	}
	align := "left-0"
	switch p.Align {
	case "right":
		align = "right-0"
	case "center":
		align = "left-1/2 -translate-x-1/2"
	}
	cls := "absolute top-full z-50 mt-1 " + width + " " + align +
		" rounded-md border border-border bg-popover shadow-md"
	nodes := []g.Node{h.Class(cls)}
	nodes = append(nodes, g.Group(p.Extra))
	nodes = append(nodes, g.Group(children))
	return h.Div(nodes...)
}
