package feedback

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// ToastVariant selects the colour of a toast.
type ToastVariant string

const (
	ToastDefault ToastVariant = "default"
	ToastSuccess ToastVariant = "success"
	ToastError   ToastVariant = "error"
	ToastWarning ToastVariant = "warning"
	ToastInfo    ToastVariant = "info"
)

// ToastPosition selects where the toast appears on screen.
type ToastPosition string

const (
	PositionTopRight     ToastPosition = "top-4 right-4"
	PositionTopLeft      ToastPosition = "top-4 left-4"
	PositionTopCenter    ToastPosition = "top-4 left-1/2 -translate-x-1/2"
	PositionBottomRight  ToastPosition = "bottom-4 right-4"
	PositionBottomLeft   ToastPosition = "bottom-4 left-4"
	PositionBottomCenter ToastPosition = "bottom-4 left-1/2 -translate-x-1/2"
)

// ToastProps configures a static server-rendered toast notification.
type ToastProps struct {
	ID            string
	Title         string
	Description   string
	Variant       ToastVariant
	Position      ToastPosition
	Dismissible   bool
	ShowIndicator bool
	Extra         []g.Node
}

func toastVariantClasses(v ToastVariant) string {
	switch v {
	case ToastSuccess:
		return "border-success/20 bg-success/10 text-success"
	case ToastError:
		return "border-destructive/20 bg-destructive/10 text-destructive"
	case ToastWarning:
		return "border-warning/20 bg-warning/10 text-warning"
	case ToastInfo:
		return "border-blue-200 bg-blue-50 text-blue-900"
	default:
		return "border-border bg-background text-foreground"
	}
}

// Toast renders a fixed-position toast notification.
// For HTMX out-of-band swaps, add hx-swap-oob="true" via Extra.
func Toast(p ToastProps) g.Node {
	pos := string(p.Position)
	if pos == "" {
		pos = string(PositionBottomRight)
	}
	cls := "fixed z-50 max-w-sm rounded-lg border p-4 shadow-md " + pos + " " + toastVariantClasses(p.Variant)
	nodes := []g.Node{h.Class(cls)}
	if p.ID != "" {
		nodes = append(nodes, h.ID(p.ID))
	}
	if p.Dismissible {
		nodes = append(nodes, g.Attr("data-dismissible", ""))
	}
	nodes = append(nodes, g.Group(p.Extra))
	if p.Title != "" {
		nodes = append(nodes, h.P(h.Class("font-medium text-sm"), g.Text(p.Title)))
	}
	if p.Description != "" {
		nodes = append(nodes, h.P(h.Class("text-sm mt-1 opacity-80"), g.Text(p.Description)))
	}
	if p.Dismissible {
		nodes = append(nodes, h.Button(
			g.Attr("data-dismiss", ""),
			h.Class("absolute top-2 right-2 opacity-50 hover:opacity-100 transition-opacity text-xs"),
			h.Type("button"),
			g.Text("×"),
		))
	}
	return h.Div(nodes...)
}
