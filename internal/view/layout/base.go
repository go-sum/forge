// Package layout provides the application's HTML shell and navigation structure.
package layout

import (
	"starter/pkg/assets"
	"starter/pkg/components/interactive"
	"starter/pkg/components/patterns/flash"
	uilayout "starter/pkg/components/ui/layout"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Props configures the full-page HTML shell.
type Props struct {
	Title           string
	CSRFToken       string
	IsAuthenticated bool
	UserName        string
	Flash           []flash.Message
	NavConfig       uilayout.NavConfig
	Children        []g.Node
}

// Page renders a complete HTML5 document with shadcn/ui theming, CSRF injection,
// flash alerts, and deferred script loading for the bundled app runtime and htmx.
func Page(p Props) g.Node {
	return h.Doctype(
		h.HTML(
			h.Lang("en"),
			h.Head(
				h.Meta(h.Charset("utf-8")),
				h.Meta(h.Name("viewport"), h.Content("width=device-width, initial-scale=1")),
				// Disable htmx inline style injection to satisfy Content-Security-Policy: style-src 'self'.
				h.Meta(h.Name("htmx-config"), h.Content(`{"includeIndicatorStyles":false}`)),
				h.TitleEl(g.Text(p.Title)),
				h.Link(h.Rel("stylesheet"), h.Href(assets.Path("css/app.css"))),
				// Prevents flash of light-mode content on dark-preference loads.
				// Must run synchronously before body paint — no defer, no src.
				interactive.ThemeScript(),
				// CSRF meta tag for non-HTMX fetch calls.
				h.Meta(h.Name("csrf-token"), h.Content(p.CSRFToken)),
			),
			h.Body(
				h.Class("bg-background text-foreground min-h-screen"),
				// Inject CSRF token into all HTMX requests via hx-headers on <body>.
				g.Attr("hx-headers", `{"X-CSRF-Token":"`+p.CSRFToken+`"}`),
				uilayout.NavMenu(uilayout.NavMenuProps{
					ID:              "app-navmenu",
					Config:          p.NavConfig,
					IsAuthenticated: p.IsAuthenticated,
					Slots:           pageNavSlots(p),
				}),
				h.Main(
					h.Class("container mx-auto px-4 py-6"),
					g.Group(p.Children),
				),
				// Toast container for flash messages and HTMX out-of-band swap notifications.
				h.Div(
					h.ID("toast-container"),
					h.Class("fixed bottom-4 right-4 z-50 flex flex-col gap-2"),
					flash.Render(p.Flash),
				),
				// app.js bundles the shared component runtime and must load before htmx.
				h.Script(h.Src(assets.Path("js/app.js")), h.Defer()),
				h.Script(h.Src(assets.Path("js/htmx.min.js")), h.Defer()),
			),
		),
	)
}

func pageNavSlots(p Props) uilayout.NavSlots {
	return uilayout.NavSlots{
		"user_name": uilayout.TextSlot(p.UserName),
		"logout": uilayout.FormSlot(uilayout.FormSlotProps{
			Label:  "Logout",
			Action: "/logout",
			HiddenFields: []uilayout.NavHiddenField{{
				Name:  "_csrf",
				Value: p.CSRFToken,
			}},
		}),
		"theme_toggle": uilayout.ControlSlot("Theme", interactive.ThemeSelector()),
	}
}
