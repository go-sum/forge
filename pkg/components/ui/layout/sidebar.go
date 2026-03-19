package layout

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// SidebarProps configures a drawer-style sidebar layout.
type SidebarProps struct {
	Nav     g.Node
	Content []g.Node
}

// Sidebar renders a responsive sidebar layout. On lg+ screens it stays open
// (static, always visible). On smaller screens it toggles via an overlay drawer.
//
// The hamburger toggle button (wherever it lives in the navbar) must carry
// data-sidebar-toggle; the backdrop div carries data-sidebar-close. Both are
// handled by the delegated click listener in static/js/app.js.
func Sidebar(p SidebarProps) g.Node {
	return h.Div(
		h.Class("flex min-h-screen"),
		// Backdrop — mobile only, hidden by default; shown when sidebar is open.
		h.Div(
			h.ID("app-sidebar-backdrop"),
			g.Attr("data-sidebar-close", ""),
			h.Class("fixed inset-0 z-20 bg-black/50 lg:hidden hidden"),
		),
		// Sidebar panel — starts off-screen on mobile; fixed open on lg+.
		h.Aside(
			h.ID("app-sidebar"),
			h.Class("fixed inset-y-0 left-0 z-30 w-64 -translate-x-full transform border-r bg-background transition-transform duration-200 lg:static lg:translate-x-0"),
			h.Nav(h.Class("flex flex-col gap-1 p-4"), p.Nav),
		),
		// Main content
		h.Div(h.Class("flex-1 overflow-auto"), g.Group(p.Content)),
	)
}
