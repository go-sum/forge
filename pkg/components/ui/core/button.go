// Package core provides fundamental UI building blocks used across all pages.
package core

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Variant selects the visual style of a button.
type Variant string

const (
	VariantDefault     Variant = "default"
	VariantDestructive Variant = "destructive"
	VariantOutline     Variant = "outline"
	VariantSecondary   Variant = "secondary"
	VariantGhost       Variant = "ghost"
	VariantLink        Variant = "link"
)

// Size selects the size of a button.
type Size string

const (
	SizeDefault Size = "default"
	SizeSm      Size = "sm"
	SizeLg      Size = "lg"
)

// ButtonProps configures a Button. Set Href to render an <a> instead of <button>.
type ButtonProps struct {
	ID      string
	Label   string
	Variant Variant
	Size    Size
	// Type defaults to "button" to avoid accidental form submission.
	Type      string
	Href      string
	Target    string
	Disabled  bool
	FullWidth bool
	// Children overrides Label for icon buttons or mixed content.
	Children []g.Node
	Extra    []g.Node
}

const baseClasses = "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] cursor-pointer"

func variantClasses(v Variant) string {
	switch v {
	case VariantDestructive:
		return "bg-destructive text-white shadow-xs hover:bg-destructive/90"
	case VariantOutline:
		return "border bg-background text-foreground shadow-xs hover:bg-accent hover:text-accent-foreground"
	case VariantSecondary:
		return "bg-secondary text-secondary-foreground shadow-xs hover:bg-secondary/80"
	case VariantGhost:
		return "hover:bg-accent hover:text-accent-foreground"
	case VariantLink:
		return "text-primary underline-offset-4 hover:underline"
	default:
		return "bg-primary text-primary-foreground shadow-xs hover:bg-primary/90"
	}
}

func sizeClasses(s Size) string {
	switch s {
	case SizeSm:
		return "h-8 rounded-md gap-1.5 px-3"
	case SizeLg:
		return "h-10 rounded-md px-6"
	default:
		return "h-9 px-4 py-2"
	}
}

func buttonClass(p ButtonProps) string {
	cls := baseClasses + " " + variantClasses(p.Variant) + " " + sizeClasses(p.Size)
	if p.FullWidth {
		cls += " w-full"
	}
	if p.Disabled {
		cls += " pointer-events-none opacity-50"
	}
	return cls
}

func buttonType(t string) string {
	if t == "" {
		return "button"
	}
	return t
}

func buttonContent(p ButtonProps) g.Node {
	if len(p.Children) > 0 {
		return g.Group(p.Children)
	}
	return g.Text(p.Label)
}

// Button renders a <button> or <a> with shadcn/ui styling.
// When Href is set, renders an <a> element (link variant).
func Button(p ButtonProps) g.Node {
	if p.Href != "" {
		nodes := []g.Node{h.Class(buttonClass(p))}
		if p.Disabled {
			nodes = append(nodes, g.Attr("aria-disabled", "true"), g.Attr("tabindex", "-1"))
		} else {
			nodes = append(nodes, h.Href(p.Href))
		}
		if p.ID != "" {
			nodes = append(nodes, h.ID(p.ID))
		}
		if p.Target != "" && !p.Disabled {
			nodes = append(nodes, h.Target(p.Target))
		}
		nodes = append(nodes, g.Group(p.Extra))
		nodes = append(nodes, buttonContent(p))
		return h.A(nodes...)
	}
	nodes := []g.Node{
		h.Class(buttonClass(p)),
		h.Type(buttonType(p.Type)),
	}
	if p.ID != "" {
		nodes = append(nodes, h.ID(p.ID))
	}
	if p.Disabled {
		nodes = append(nodes, h.Disabled())
	}
	nodes = append(nodes, g.Group(p.Extra))
	nodes = append(nodes, buttonContent(p))
	return h.Button(nodes...)
}
