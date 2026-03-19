// Package dropdown provides a native HTML dropdown menu using <details>/<summary>.
// Open/close is handled natively by the browser. A delegated outside-click
// listener in static/js/app.js closes any open [data-dropdown] details element
// when the user clicks outside of it.
package dropdown

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Props configures a dropdown root.
type Props struct {
	ID    string
	Extra []g.Node
}

// Root renders a dropdown root as a native <details> element.
// data-dropdown is the hook for the outside-click delegated listener in app.js.
func Root(p Props, children ...g.Node) g.Node {
	nodes := []g.Node{
		g.Attr("data-dropdown", ""),
		h.Class("relative inline-block"),
	}
	if p.ID != "" {
		nodes = append(nodes, h.ID(p.ID))
	}
	nodes = append(nodes, g.Group(p.Extra))
	nodes = append(nodes, g.Group(children))
	return h.Details(nodes...)
}

// Trigger renders the <summary> click trigger.
func Trigger(children ...g.Node) g.Node {
	return h.Summary(
		h.Class("list-none cursor-pointer"),
		g.Group(children),
	)
}

// Content renders the dropdown panel, visible when <details> is open.
func Content(children ...g.Node) g.Node {
	return h.Div(
		h.Class("absolute z-50 mt-1 min-w-[8rem] rounded-md border border-border bg-popover p-1 shadow-md"),
		g.Group(children),
	)
}

// Item renders a single menu entry as <a> (href set) or <button>.
func Item(label, href string, disabled bool) g.Node {
	cls := "flex w-full items-center px-2 py-1.5 text-sm rounded-sm transition-colors"
	if disabled {
		cls += " opacity-50 pointer-events-none"
	} else {
		cls += " hover:bg-accent hover:text-accent-foreground cursor-default"
	}
	if href != "" {
		return h.A(h.Class(cls), h.Href(href), g.Text(label))
	}
	return h.Button(
		h.Class(cls),
		h.Type("button"),
		g.If(disabled, h.Disabled()),
		g.Text(label),
	)
}

// Separator renders a horizontal rule between menu sections.
func Separator() g.Node {
	return h.Div(h.Class("h-px my-1 -mx-1 bg-muted"), h.Role("separator"))
}

// Label renders a non-interactive section heading.
func Label(label string) g.Node {
	return h.Div(h.Class("px-2 py-1.5 text-sm font-semibold"), g.Text(label))
}
