// Package tooltip provides a CSS-first tooltip with ARIA linkage helpers.
// Two variants are available:
//
//   - Hover variant (Root/Trigger/Content): shown on CSS hover and focus-within.
//   - Click variant (ClickRoot/ClickTrigger/ClickContent): shown on click via
//     native <details>/<summary> toggle. Useful for touch/mobile where hover
//     is unavailable. Shares the data-popover outside-click listener in static/js/components.js.
package tooltip

import (
	core "starter/pkg/components/ui/core"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// TriggerAttrs returns the ARIA attributes the focusable trigger element should carry.
func TriggerAttrs(id string) []g.Node {
	return []g.Node{g.Attr("aria-describedby", id)}
}

// ── Hover variant ────────────────────────────────────────────────────────────

// Root renders the hover-tooltip root container.
func Root(children ...g.Node) g.Node {
	return h.Div(
		h.Class("group relative inline-flex"),
		g.Group(children),
	)
}

// Trigger renders a wrapper around the trigger element so Root can react to hover and focus-within.
func Trigger(children ...g.Node) g.Node {
	return h.Span(
		h.Class("inline-flex"),
		g.Group(children),
	)
}

// Content renders the tooltip panel. It is shown on hover and when the trigger receives focus.
func Content(id string, children ...g.Node) g.Node {
	return h.Div(
		h.ID(id),
		h.Role("tooltip"),
		h.Class("absolute bottom-full left-1/2 z-50 mb-2 hidden -translate-x-1/2 whitespace-nowrap rounded-md border bg-popover px-3 py-1.5 text-xs text-popover-foreground shadow-md pointer-events-none group-hover:block group-focus-within:block"),
		g.Group(children),
	)
}

// ── Click variant ────────────────────────────────────────────────────────────

// ClickRoot renders a click-activated tooltip using core.Popover.Root.
// data-popover hooks the shared outside-click listener in static/js/components.js.
func ClickRoot(children ...g.Node) g.Node {
	return core.Popover.Root(core.PopoverRootProps{
		Class: "relative inline-flex",
	}, children...)
}

// ClickTrigger renders a <summary> via core.Popover.Trigger that toggles the
// click tooltip. Pass aria-describedby via TriggerAttrs — it belongs on the
// <summary> since that is the interactive element in the click variant.
func ClickTrigger(children ...g.Node) g.Node {
	return core.Popover.Trigger(core.PopoverTriggerProps{
		Class: "inline-flex",
	}, children...)
}

// ClickContent renders the click-tooltip panel above the trigger.
// Visibility is handled natively by the browser's <details> UA stylesheet —
// no hidden/group-open classes are needed.
func ClickContent(id string, children ...g.Node) g.Node {
	return h.Div(
		h.ID(id),
		h.Role("tooltip"),
		h.Class("absolute bottom-full left-1/2 z-50 mb-2 -translate-x-1/2 whitespace-nowrap rounded-md border bg-popover px-3 py-1.5 text-xs text-popover-foreground shadow-md"),
		g.Group(children),
	)
}
