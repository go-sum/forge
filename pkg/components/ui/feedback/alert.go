// Package feedback provides notification and status-indicator components.
package feedback

import (
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	componenticons "starter/pkg/components/icons"
	iconrender "starter/pkg/components/icons/render"
	core "starter/pkg/components/ui/core"
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
	Extra       []g.Node
}

func alertVariantClasses(v AlertVariant) string {
	if v == AlertDestructive {
		return "text-destructive bg-card"
	}
	return "bg-card text-card-foreground"
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
		h.Class(cls),
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
// delegated data-dismiss handler in js/app.js.
func (alertNS) Root(p AlertProps, children ...g.Node) g.Node {
	cls := "relative w-full rounded-lg border px-4 py-3 text-sm grid grid-cols-[0_1fr] gap-y-0.5 items-start " + alertVariantClasses(p.Variant)
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
	nodes = append(nodes, g.Group(children))
	if p.Dismissible {
		nodes = append(nodes, dismissButton("absolute top-3 right-3 opacity-70 hover:opacity-100 transition-opacity"))
	}
	return h.Div(nodes...)
}

// Title renders the alert heading.
func (alertNS) Title(children ...g.Node) g.Node {
	return h.H5(
		h.Class("col-start-2 line-clamp-1 min-h-4 font-medium tracking-tight"),
		g.Group(children),
	)
}

// Description renders the alert body text.
func (alertNS) Description(children ...g.Node) g.Node {
	return h.Div(
		h.Class("text-muted-foreground col-start-2 grid justify-items-start gap-1 text-sm"),
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
