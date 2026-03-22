// Package feedback provides notification and status-indicator components.
package feedback

import (
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	componenticons "github.com/go-sum/componentry/icons"
	iconrender "github.com/go-sum/componentry/icons/render"
	core "github.com/go-sum/componentry/ui/core"
)

// AlertVariant selects the visual style of an alert.
type AlertVariant string

const (
	AlertDefault     AlertVariant = "default"
	AlertDestructive AlertVariant = "destructive"
)

// AlertProps configures a single alert banner.
type AlertProps struct {
	ID          string
	Variant     AlertVariant
	Dismissible bool
	// Icon is an optional leading icon node (e.g. core.Icon(...)).
	// When set, the layout switches to a two-column grid so the icon sits in
	// its own column alongside the title/description — no call-site changes needed.
	Icon  g.Node
	Extra []g.Node
}

func alertVariantClasses(v AlertVariant) string {
	if v == AlertDestructive {
		return "backdrop-blur-sm border-destructive/30 bg-destructive/20 text-destructive [&_[data-alert-description]]:text-destructive/80"
	}
	return "backdrop-blur-sm border-primary/30 bg-primary/20 text-primary [&_[data-alert-description]]:text-muted-foreground"
}

func alertVariantForType(kind string) AlertVariant {
	switch strings.ToLower(kind) {
	case "destructive", "error":
		return AlertDestructive
	default:
		return AlertDefault
	}
}

// dismissButton renders a shared dismiss <button> used by Alert and Toast.
func dismissButton(cls string) g.Node {
	return h.Button(
		g.Attr("data-dismiss", ""),
		h.Class(cls+" outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"),
		h.Type("button"),
		g.Attr("aria-label", "Dismiss"),
		core.Icon(iconrender.PropsFor(componenticons.Close, core.IconProps{Size: "size-4"})),
	)
}

type alertNS struct{}

// Alert groups alert sub-components under a namespace: Alert.Root, Alert.Title, Alert.Description, Alert.List.
var Alert alertNS

// Root renders a shadcn/ui-style alert. When Dismissible is true, a close
// button is added; clicking it removes the element from the DOM via the
// delegated data-dismiss handler in static/js/components.js.
// When Icon is set, the layout switches to a two-column grid: the icon
// occupies the first column and the children are wrapped in a div for the second.
func (alertNS) Root(p AlertProps, children ...g.Node) g.Node {
	var cls string
	if p.Icon != nil {
		cls = "relative w-full rounded-lg border px-4 py-3 text-sm grid grid-cols-[auto_1fr] gap-x-3 items-start " + alertVariantClasses(p.Variant)
	} else {
		cls = "relative w-full rounded-lg border px-4 py-3 text-sm grid gap-1.5 items-start " + alertVariantClasses(p.Variant)
	}
	nodes := []g.Node{
		h.Class(cls),
		h.Role("alert"),
	}
	if p.ID != "" {
		nodes = append(nodes, h.ID(p.ID))
	}
	if p.Dismissible {
		nodes = append(nodes, g.Attr("data-dismissible", ""))
	}
	nodes = append(nodes, g.Group(p.Extra))
	if p.Icon != nil {
		nodes = append(nodes,
			h.Div(g.Attr("data-alert-icon", ""), h.Class("mt-0.5"), p.Icon),
			h.Div(h.Class("grid gap-1.5"), g.Group(children)),
		)
	} else {
		nodes = append(nodes, g.Group(children))
	}
	if p.Dismissible {
		nodes = append(nodes, dismissButton("absolute top-3 right-3 opacity-70 hover:opacity-100 transition-opacity"))
	}
	return h.Div(nodes...)
}

// Title renders the alert heading.
func (alertNS) Title(children ...g.Node) g.Node {
	return h.H5(
		h.Class("line-clamp-1 min-h-4 font-medium tracking-tight"),
		g.Group(children),
	)
}

// Description renders the alert body text.
func (alertNS) Description(children ...g.Node) g.Node {
	return h.Div(
		h.Class("grid justify-items-start gap-1 text-sm"),
		g.Attr("data-alert-description", ""),
		g.Group(children),
	)
}

// List renders multiple dismissible alerts from parallel type/text slices.
// Non-destructive types fall back to AlertDefault because Alert only exposes
// default and destructive variants.
func (alertNS) List(types []string, texts []string) g.Node {
	n := min(len(texts), len(types))
	nodes := make([]g.Node, n)
	for i := range n {
		nodes[i] = Alert.Root(
			AlertProps{Variant: alertVariantForType(types[i]), Dismissible: true},
			Alert.Description(g.Text(texts[i])),
		)
	}
	return g.Group(nodes)
}
