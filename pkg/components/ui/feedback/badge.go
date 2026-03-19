package feedback

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// BadgeVariant selects the visual style of a badge.
type BadgeVariant string

const (
	BadgeDefault     BadgeVariant = "default"
	BadgeSecondary   BadgeVariant = "secondary"
	BadgeDestructive BadgeVariant = "destructive"
	BadgeOutline     BadgeVariant = "outline"
)

// BadgeProps configures a status badge.
type BadgeProps struct {
	ID       string
	Variant  BadgeVariant
	Children []g.Node
	Extra    []g.Node
}

func badgeVariantClasses(v BadgeVariant) string {
	switch v {
	case BadgeDestructive:
		return "border-transparent bg-destructive text-white"
	case BadgeOutline:
		return "text-foreground"
	case BadgeSecondary:
		return "border-transparent bg-secondary text-secondary-foreground"
	default:
		return "border-transparent bg-primary text-primary-foreground"
	}
}

// Badge renders a small status indicator <span>.
func Badge(p BadgeProps) g.Node {
	cls := "inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap shrink-0 transition-colors " + badgeVariantClasses(p.Variant)
	nodes := []g.Node{h.Class(cls)}
	if p.ID != "" {
		nodes = append(nodes, h.ID(p.ID))
	}
	nodes = append(nodes, g.Group(p.Extra))
	nodes = append(nodes, g.Group(p.Children))
	return h.Span(nodes...)
}
