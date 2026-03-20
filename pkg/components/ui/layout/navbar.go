// Package layout provides structural shell components for page layout.
package layout

import (
	componenticons "starter/pkg/components/icons"
	iconrender "starter/pkg/components/icons/render"
	"starter/pkg/components/ui/core"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// NavConfig is the top-level navigation configuration parsed from nav.yaml.
type NavConfig struct {
	Brand    NavBrand     `koanf:"brand"`
	Sections []NavSection `koanf:"sections"`
}

// NavBrand configures the logo/wordmark shown at the start of the nav bar.
type NavBrand struct {
	Label    string `koanf:"label"`
	Href     string `koanf:"href"`
	LogoPath string `koanf:"logo_path"`
}

// NavSection is a named group of NavItems within the navigation menu.
type NavSection struct {
	Label string    `koanf:"label"`
	Items []NavItem `koanf:"items"`
}

// NavItem models both leaf links and recursive menu parents.
// Special item types render built-in controls such as separators, theme toggle,
// user name, and logout.
type NavItem struct {
	Type       string    `koanf:"type" validate:"omitempty,oneof=separator user_name logout theme_toggle"`
	Visibility string    `koanf:"visibility" validate:"omitempty,oneof=all guest user"`
	Label      string    `koanf:"label"`
	Href       string    `koanf:"href"`
	Items      []NavItem `koanf:"items"`
}

func (i NavItem) IsSeparator() bool {
	return i.Type == "separator"
}

func (i NavItem) HasChildren() bool {
	return len(i.Items) > 0
}

const defaultNavMenuID = "navmenu"

// NavMenuProps configures a responsive navigation menu.
type NavMenuProps struct {
	ID              string
	Config          NavConfig
	CSRFToken       string
	IsAuthenticated bool
	UserName        string
	// ThemeSelector is the rendered theme-toggle node injected by the caller.
	// When nil, theme_toggle items render nothing.
	ThemeSelector g.Node
}

func navMenuID(id string) string {
	if id == "" {
		return defaultNavMenuID
	}
	return id
}

// NavMenu renders a CSS-only responsive navigation menu. Desktop uses horizontal
// sections with native <details> dropdowns; mobile uses a drawer sidebar.
func NavMenu(p NavMenuProps) g.Node {
	drawerID := navMenuID(p.ID)

	return h.Nav(
		h.Class("w-full border-b bg-background"),
		h.Div(
			h.Class("container mx-auto flex h-14 items-center px-4"),
			h.Div(h.Class("mr-4 flex shrink-0"), brandNode(p.Config.Brand)),
			h.Div(
				h.Class("hidden min-w-0 flex-1 items-center md:flex"),
				desktopNavRegion(p),
			),
			mobileToggleButton(drawerID),
		),
		h.Div(
			h.Class("md:hidden"),
			Sidebar(SidebarProps{
				ID:  drawerID,
				Nav: mobileDrawer(drawerID, p),
			}),
		),
	)
}

func brandNode(brand NavBrand) g.Node {
	label := brand.Label
	if label == "" {
		label = "Starter"
	}
	href := brand.Href
	if href == "" {
		href = "/"
	}

	children := []g.Node{}
	if brand.LogoPath != "" {
		children = append(children, h.Img(
			h.Src(brand.LogoPath),
			h.Alt(label),
			h.Class("h-8 w-8 rounded-md object-contain"),
		))
	}
	children = append(children, h.Span(h.Class("truncate"), g.Text(label)))

	return h.A(
		h.Href(href),
		h.Class("flex shrink-0 items-center gap-3 text-lg font-semibold tracking-tight"),
		g.Group(children),
	)
}

func mobileToggleButton(id string) g.Node {
	nodes := append([]g.Node{
		h.Class("ml-auto inline-flex items-center justify-center rounded-md p-2 text-foreground transition-colors hover:bg-accent hover:text-accent-foreground md:hidden"),
	}, ToggleAttrs(id)...)
	nodes = append(nodes,
		h.Span(h.Class("sr-only"), g.Text("Open navigation menu")),
		h.Span(
			h.Class("inline-flex h-4 w-5 flex-col justify-between"),
			h.Span(h.Class("block h-0.5 w-full rounded-full bg-current")),
			h.Span(h.Class("block h-0.5 w-full rounded-full bg-current")),
			h.Span(h.Class("block h-0.5 w-full rounded-full bg-current")),
		),
	)
	return h.Label(nodes...)
}

func desktopNavRegion(p NavMenuProps) g.Node {
	sections := renderDesktopSections(p)
	if len(sections) == 0 {
		return h.Div(h.Class("flex flex-1 items-center gap-2"))
	}
	return h.Div(
		h.Class("flex min-w-0 flex-1 items-center justify-between gap-6"),
		g.Group(sections),
	)
}

func mobileDrawer(id string, p NavMenuProps) g.Node {
	sections := renderMobileSections(p)
	return h.Div(
		h.Class("flex h-full flex-col"),
		h.Div(
			h.Class("flex items-center justify-between border-b border-border px-4 py-4"),
			brandNode(p.Config.Brand),
			mobileCloseButton(id),
		),
		h.Div(
			h.Class("flex min-h-0 flex-1 flex-col justify-between gap-8 overflow-y-auto px-4 py-5"),
			h.Div(
				h.Class("flex flex-1 flex-col justify-between gap-8"),
				g.Group(sections),
			),
		),
	)
}

func mobileCloseButton(id string) g.Node {
	nodes := append([]g.Node{
		h.Class("inline-flex items-center justify-center rounded-md p-2 text-foreground transition-colors hover:bg-accent hover:text-accent-foreground"),
	}, CloseAttrs(id)...)
	nodes = append(nodes,
		h.Span(h.Class("sr-only"), g.Text("Close navigation menu")),
		h.Span(
			h.Class("relative block size-4"),
			h.Span(h.Class("absolute left-1/2 top-1/2 block h-0.5 w-4 -translate-x-1/2 -translate-y-1/2 rotate-45 rounded-full bg-current")),
			h.Span(h.Class("absolute left-1/2 top-1/2 block h-0.5 w-4 -translate-x-1/2 -translate-y-1/2 -rotate-45 rounded-full bg-current")),
		),
	)
	return h.Label(nodes...)
}

func renderDesktopSections(p NavMenuProps) []g.Node {
	nodes := []g.Node{}
	for _, section := range p.Config.Sections {
		node := renderDesktopSection(section, p)
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func renderDesktopSection(section NavSection, p NavMenuProps) g.Node {
	items := []g.Node{}
	for _, item := range section.Items {
		node := renderDesktopItem(item, p, 0)
		if node != nil {
			items = append(items, node)
		}
	}
	if len(items) == 0 && section.Label == "" {
		return nil
	}

	children := []g.Node{}
	if section.Label != "" {
		children = append(children, h.Span(
			h.Class("text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground"),
			g.Text(section.Label),
		))
	}
	children = append(children, h.Div(h.Class("flex items-stretch gap-0"), g.Group(items)))

	return h.Div(h.Class("flex min-w-0 items-center gap-3"), g.Group(children))
}

func renderDesktopItem(item NavItem, p NavMenuProps, level int) g.Node {
	if !itemVisible(item, p.IsAuthenticated) {
		return nil
	}

	switch {
	case item.IsSeparator():
		if level == 0 {
			return nil
		}
		return h.Div(h.Class("my-2 h-px bg-border"), h.Role("separator"))
	case item.HasChildren():
		return renderDesktopParent(item, p, level)
	case item.Type != "":
		return renderDesktopSpecialItem(item, p)
	case item.Label == "" || item.Href == "":
		return nil
	case level == 0:
		return h.A(
			h.Href(item.Href),
			h.Class("inline-flex items-center px-4 py-3 text-sm font-medium text-foreground transition-colors hover:bg-accent/60 hover:text-accent-foreground"),
			g.Text(item.Label),
		)
	default:
		return h.A(
			h.Href(item.Href),
			h.Class("block w-full px-4 py-3 text-sm transition-colors hover:bg-accent/60 hover:text-accent-foreground"),
			g.Text(item.Label),
		)
	}
}

func renderDesktopParent(item NavItem, p NavMenuProps, level int) g.Node {
	if item.Label == "" {
		return nil
	}

	nodes := desktopSubmenuNodes(item, p, level+1)
	if len(nodes) == 0 && item.Href == "" {
		return nil
	}

	summaryClass := "navmenu-summary flex list-none cursor-pointer items-center text-sm transition-colors hover:bg-accent/60 hover:text-accent-foreground"
	panelClass := "absolute z-50 min-w-[16rem] border border-border bg-popover shadow-lg"
	icon := core.Icon(iconrender.PropsFor(componenticons.ChevronDown, core.IconProps{
		Size: "navmenu-chevron size-4 shrink-0 text-muted-foreground transition-transform",
	}))

	if level == 0 {
		return h.Details(
			h.Class("group relative"),
			h.Summary(
				h.Class(summaryClass+" gap-2 px-4 py-3 font-medium"),
				h.Span(g.Text(item.Label)),
				icon,
			),
			h.Div(
				h.Class(panelClass+" left-0 top-full mt-px flex flex-col divide-y divide-border"),
				g.Group(nodes),
			),
		)
	}

	return h.Details(
		h.Class("border-t border-border/70 bg-background"),
		h.Summary(
			h.Class("navmenu-summary flex list-none items-center justify-between gap-3 px-4 py-3 text-sm font-medium transition-colors hover:bg-accent/60 hover:text-accent-foreground"),
			h.Span(g.Text(item.Label)),
			core.Icon(iconrender.PropsFor(componenticons.ChevronDown, core.IconProps{
				Size: "navmenu-chevron size-4 shrink-0 text-muted-foreground transition-transform",
			})),
		),
		h.Div(
			h.Class("flex flex-col divide-y divide-border border-t border-border/70"),
			g.Group(nodes),
		),
	)
}

func renderDesktopSpecialItem(item NavItem, p NavMenuProps) g.Node {
	switch item.Type {
	case "theme_toggle":
		return p.ThemeSelector
	case "user_name":
		if !p.IsAuthenticated || p.UserName == "" {
			return nil
		}
		return h.Span(h.Class("text-sm text-muted-foreground"), g.Text(p.UserName))
	case "logout":
		if !p.IsAuthenticated {
			return nil
		}
		label := item.Label
		if label == "" {
			label = "Logout"
		}
		return h.Form(
			h.Method("post"),
			h.Action("/logout"),
			h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(p.CSRFToken)),
			core.Button(core.ButtonProps{Label: label, Variant: core.VariantGhost, Size: core.SizeSm, Type: "submit"}),
		)
	default:
		return nil
	}
}

func desktopSubmenuNodes(item NavItem, p NavMenuProps, level int) []g.Node {
	nodes := []g.Node{}
	if item.Href != "" && item.Label != "" {
		nodes = append(nodes, h.A(
			h.Href(item.Href),
			h.Class("block w-full px-4 py-4 text-sm font-medium transition-colors hover:bg-accent/60 hover:text-accent-foreground"),
			g.Text(item.Label),
		))
	}
	childNodes := []g.Node{}
	for _, child := range item.Items {
		node := renderDesktopItem(child, p, level)
		if node != nil {
			childNodes = append(childNodes, node)
		}
	}
	if len(nodes) > 0 && len(childNodes) > 0 {
		nodes = append(nodes, h.Div(h.Class("my-2 h-px bg-border"), h.Role("separator")))
	}
	return append(nodes, childNodes...)
}

func renderMobileSections(p NavMenuProps) []g.Node {
	nodes := []g.Node{}
	for _, section := range p.Config.Sections {
		node := renderMobileSection(section, p)
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func renderMobileSection(section NavSection, p NavMenuProps) g.Node {
	items := []g.Node{}
	for _, item := range section.Items {
		node := renderMobileItem(item, p, 0)
		if node != nil {
			items = append(items, node)
		}
	}
	if len(items) == 0 && section.Label == "" {
		return nil
	}

	children := []g.Node{}
	if section.Label != "" {
		children = append(children, h.P(
			h.Class("px-1 text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground"),
			g.Text(section.Label),
		))
	}
	children = append(children, h.Div(h.Class("w-full divide-y divide-border border-y border-border"), g.Group(items)))

	return h.Div(h.Class("flex flex-col gap-2"), g.Group(children))
}

func renderMobileItem(item NavItem, p NavMenuProps, level int) g.Node {
	if !itemVisible(item, p.IsAuthenticated) {
		return nil
	}

	switch {
	case item.IsSeparator():
		return h.Div(h.Class("mx-2 my-2 h-px bg-border"), h.Role("separator"))
	case item.HasChildren():
		return renderMobileParent(item, p, level)
	case item.Type != "":
		return renderMobileSpecialItem(item, p)
	case item.Label == "" || item.Href == "":
		return nil
	default:
		return h.A(
			h.Href(item.Href),
			h.Class("block w-full px-4 py-4 text-sm font-medium transition-colors hover:bg-accent/60 hover:text-accent-foreground"),
			g.Text(item.Label),
		)
	}
}

func renderMobileParent(item NavItem, p NavMenuProps, level int) g.Node {
	if item.Label == "" {
		return nil
	}

	nodes := mobileSubmenuNodes(item, p, level+1)
	if len(nodes) == 0 && item.Href == "" {
		return nil
	}

	return h.Details(
		h.Class("bg-background"),
		h.Summary(
			h.Class("navmenu-summary flex list-none cursor-pointer items-center justify-between gap-3 px-4 py-4 text-left text-sm font-medium transition-colors hover:bg-accent/60 hover:text-accent-foreground"),
			h.Span(g.Text(item.Label)),
			core.Icon(iconrender.PropsFor(componenticons.ChevronDown, core.IconProps{
				Size: "navmenu-chevron size-4 shrink-0 text-muted-foreground transition-transform",
			})),
		),
		h.Div(
			h.Class("flex flex-col divide-y divide-border border-t border-border"),
			g.Group(nodes),
		),
	)
}

func renderMobileSpecialItem(item NavItem, p NavMenuProps) g.Node {
	switch item.Type {
	case "theme_toggle":
		return h.Div(
			h.Class("flex items-center justify-between px-4 py-4 transition-colors hover:bg-accent/60"),
			h.Span(h.Class("text-sm text-muted-foreground"), g.Text("Theme")),
			p.ThemeSelector,
		)
	case "user_name":
		if !p.IsAuthenticated || p.UserName == "" {
			return nil
		}
		return h.Span(h.Class("block px-4 py-4 text-sm font-medium text-foreground"), g.Text(p.UserName))
	case "logout":
		if !p.IsAuthenticated {
			return nil
		}
		label := item.Label
		if label == "" {
			label = "Logout"
		}
		return h.Form(
			h.Method("post"),
			h.Action("/logout"),
			h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(p.CSRFToken)),
			core.Button(core.ButtonProps{Label: label, Variant: core.VariantGhost, Size: core.SizeSm, Type: "submit", FullWidth: true}),
		)
	default:
		return nil
	}
}

func mobileSubmenuNodes(item NavItem, p NavMenuProps, level int) []g.Node {
	nodes := []g.Node{}
	if item.Href != "" && item.Label != "" {
		nodes = append(nodes, h.A(
			h.Href(item.Href),
			h.Class("block w-full px-4 py-4 text-sm font-medium transition-colors hover:bg-accent/60 hover:text-accent-foreground"),
			g.Text(item.Label),
		))
	}
	childNodes := []g.Node{}
	for _, child := range item.Items {
		node := renderMobileItem(child, p, level)
		if node != nil {
			childNodes = append(childNodes, node)
		}
	}
	if len(nodes) > 0 && len(childNodes) > 0 {
		nodes = append(nodes, h.Div(h.Class("mx-2 my-2 h-px bg-border"), h.Role("separator")))
	}
	return append(nodes, childNodes...)
}

func itemVisible(item NavItem, isAuthenticated bool) bool {
	switch item.Visibility {
	case "guest":
		return !isAuthenticated
	case "user":
		return isAuthenticated
	default:
		return true
	}
}
