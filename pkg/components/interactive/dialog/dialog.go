// Package dialog provides a native HTML <dialog> modal component.
// Trigger and content are linked by a shared ID string.
// The delegated click handler in static/js/app.js calls showModal()/close()
// in response to data-dialog-open and data-dialog-close attributes.
// Native <dialog> provides: focus trap, ESC-to-close, aria-modal, and backdrop.
package dialog

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Root is a fragment wrapper; callers place Trigger and Content
// as siblings anywhere in the tree — they are linked by dialogID, not by DOM nesting.
func Root(children ...g.Node) g.Node {
	return h.Div(h.Class("contents"), g.Group(children))
}

// Trigger renders a wrapper that opens the <dialog> with the given ID
// when clicked. The delegated handler in app.js calls showModal().
func Trigger(dialogID string, children ...g.Node) g.Node {
	return h.Div(
		g.Attr("data-dialog-open", dialogID),
		h.Class("contents cursor-pointer"),
		g.Group(children),
	)
}

// Content renders a native <dialog> element with the given ID.
// Backdrop styling is provided by `dialog::backdrop` in tailwind.css.
func Content(id string, children ...g.Node) g.Node {
	return h.Dialog(
		h.ID(id),
		h.Class("w-full max-w-lg rounded-lg border bg-background p-6 shadow-lg backdrop:bg-black/50"),
		g.Group(children),
	)
}

// Header renders the dialog header container.
func Header(children ...g.Node) g.Node {
	return h.Div(
		h.Class("flex flex-col gap-2 text-center sm:text-left mb-4"),
		g.Group(children),
	)
}

// Footer renders the dialog footer container.
func Footer(children ...g.Node) g.Node {
	return h.Div(
		h.Class("flex flex-col-reverse gap-2 sm:flex-row sm:justify-end mt-4"),
		g.Group(children),
	)
}

// Title renders the dialog heading.
func Title(children ...g.Node) g.Node {
	return h.H2(
		h.Class("text-lg leading-none font-semibold"),
		g.Group(children),
	)
}

// Description renders a muted description paragraph.
func Description(children ...g.Node) g.Node {
	return h.P(
		h.Class("text-muted-foreground text-sm"),
		g.Group(children),
	)
}

// Close renders a wrapper that closes the nearest parent <dialog> on click.
// The delegated handler in app.js calls dialog.close().
func Close(children ...g.Node) g.Node {
	return h.Div(
		g.Attr("data-dialog-close", ""),
		h.Class("contents cursor-pointer"),
		g.Group(children),
	)
}
