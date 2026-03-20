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
	PositionTopRight     ToastPosition = "top-right"
	PositionTopLeft      ToastPosition = "top-left"
	PositionTopCenter    ToastPosition = "top-center"
	PositionBottomRight  ToastPosition = "bottom-right"
	PositionBottomLeft   ToastPosition = "bottom-left"
	PositionBottomCenter ToastPosition = "bottom-center"
)

// ToastProps configures a static server-rendered toast notification.
type ToastProps struct {
	ID          string
	Title       string
	Description string
	Variant     ToastVariant
	Position    ToastPosition
	Dismissible bool
	Extra       []g.Node
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

func toastAnnouncementAttrs(v ToastVariant) []g.Node {
	if v == ToastError || v == ToastWarning {
		return []g.Node{
			h.Role("alert"),
			g.Attr("aria-live", "assertive"),
			g.Attr("aria-atomic", "true"),
		}
	}
	return []g.Node{
		h.Role("status"),
		g.Attr("aria-live", "polite"),
		g.Attr("aria-atomic", "true"),
	}
}

func toastPositionClasses(p ToastPosition) string {
	switch p {
	case PositionTopRight:
		return "top-4 right-4"
	case PositionTopLeft:
		return "top-4 left-4"
	case PositionTopCenter:
		return "top-4 left-1/2 -translate-x-1/2"
	case PositionBottomLeft:
		return "bottom-4 left-4"
	case PositionBottomCenter:
		return "bottom-4 left-1/2 -translate-x-1/2"
	case PositionBottomRight:
		return "bottom-4 right-4"
	default:
		return ""
	}
}

// Toast renders a toast notification. When Position is set, the toast is
// fixed-positioned and self-contained. When Position is "" (zero value), the
// toast renders as a plain card suitable for injection into a container div.
// For HTMX out-of-band swaps, add hx-swap-oob="true" via Extra.
func Toast(p ToastProps) g.Node {
	cardCls := "rounded-lg border p-4 shadow-md " + toastVariantClasses(p.Variant)
	var fixedCls string
	if p.Position != "" {
		fixedCls = "fixed z-50 max-w-sm " + toastPositionClasses(p.Position) + " "
	}
	cls := fixedCls + cardCls
	nodes := []g.Node{h.Class(cls)}
	nodes = append(nodes, toastAnnouncementAttrs(p.Variant)...)
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
		nodes = append(nodes, dismissButton("absolute top-2 right-2 opacity-50 hover:opacity-100 transition-opacity"))
	}
	return h.Div(nodes...)
}
