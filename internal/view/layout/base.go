// Package layout provides the application's HTML shell and navigation structure.
package layout

import (
	"github.com/go-sum/componentry/assets"
	"github.com/go-sum/componentry/interactive"
	"github.com/go-sum/componentry/patterns/flash"
	"github.com/go-sum/componentry/patterns/font"
	"github.com/go-sum/componentry/patterns/head"
	uilayout "github.com/go-sum/componentry/ui/layout"
	"github.com/go-sum/forge/config"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Props configures the full-page HTML shell.
type Props struct {
	Title           string
	FaviconPath     string
	Description     string
	MetaKeywords    []string
	OGImage         string
	CSRFFieldName   string
	CSRFHeaderName  string
	CurrentPath     string
	CSRFToken       string
	IsAuthenticated bool
	UserName        string
	Flash           []flash.Message
	NavConfig       config.NavConfig
	FontConfig      font.Config
	SignoutPath      string
	CopyrightYear   int
	AppVersion      string
	Children        []g.Node
}

// Page renders a complete HTML5 document with shadcn/ui theming, CSRF injection,
// flash alerts, and deferred script loading for the bundled app runtime and htmx.
func Page(p Props) g.Node {
	return h.Doctype(
		h.HTML(
			h.Lang("en"),
			head.Head(head.Props{
				Meta: head.MetaProps{
					Title:       p.Title,
					FaviconHref: p.FaviconPath,
					Description: p.Description,
					Keywords:    p.MetaKeywords,
					OGImage:     p.OGImage,
				},
				Stylesheets: []head.Stylesheet{{
					Href: assets.Path("css/app.css"),
				}},
				Extra: buildHeadExtras(p),
				Scripts: []head.Script{
					{Src: assets.Path("js/app.js"), Defer: true},
					{Src: assets.Path("js/htmx.min.js"), Defer: true},
				},
			}),
			h.Body(
				h.Class("bg-background text-foreground min-h-screen flex flex-col"),
				// Inject CSRF token into all HTMX requests via hx-headers on <body>.
				g.Attr("hx-headers", `{"`+p.CSRFHeaderName+`":"`+p.CSRFToken+`"}`),
				uilayout.NavMenu(uilayout.NavMenuProps{
					ID:              "app-navmenu",
					Config:          p.NavConfig,
					CurrentPath:     p.CurrentPath,
					IsAuthenticated: p.IsAuthenticated,
					Slots:           pageNavSlots(p),
				}),
				h.Main(
					h.Class("container mx-auto px-4 py-6 flex-1"),
					g.Group(p.Children),
				),
				// Toast container for flash messages and HTMX out-of-band swap notifications.
				h.Div(
					h.ID("toast-container"),
					h.Class("fixed bottom-4 right-4 z-50 flex flex-col gap-2"),
					flash.Render(p.Flash),
				),
				Footer(p),
			),
		),
	)
}

// buildHeadExtras assembles all <head> extra nodes: CSP/HTMX config, theme
// script, CSRF meta tag, and any font provider nodes derived from FontConfig.
func buildHeadExtras(p Props) []g.Node {
	extras := []g.Node{
		// Disable htmx inline style injection to satisfy Content-Security-Policy: style-src 'self'.
		h.Meta(h.Name("htmx-config"), h.Content(`{"includeIndicatorStyles":false}`)),
		// Prevents flash of light-mode content on dark-preference loads.
		// Must run synchronously before body paint — no defer, no src.
		interactive.ThemeScript(),
		// CSRF meta tag for non-HTMX fetch calls.
		h.Meta(h.Name("csrf-token"), h.Content(p.CSRFToken)),
	}
	extras = append(extras, font.Nodes(font.BuildProviders(p.FontConfig, assets.Path)...)...)
	return extras
}

func pageNavSlots(p Props) uilayout.NavSlots {
	userName := p.UserName
	if p.IsAuthenticated && userName == "" {
		userName = "Account"
	}

	return uilayout.NavSlots{
		"user_name": uilayout.TextSlot(userName),
		"signout": uilayout.FormSlot(uilayout.FormSlotProps{
			Label:  "Signout",
			Action: p.SignoutPath,
			HiddenFields: []uilayout.NavHiddenField{{
				Name:  p.CSRFFieldName,
				Value: p.CSRFToken,
			}},
		}),
		"theme_toggle": uilayout.ControlSlot("Theme", interactive.ThemeSelector()),
	}
}
