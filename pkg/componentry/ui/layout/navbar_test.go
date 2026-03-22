package layout

import (
	"strings"
	"testing"

	testutil "github.com/go-sum/componentry/testutil"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
	componenticons "github.com/go-sum/componentry/icons"
)

func TestNavbarRendersGuestItemsAndHidesUserActions(t *testing.T) {
	got := testutil.RenderNode(t, testNavbar(false, "", nil))

	checks := []string{
		`class="w-full border-b bg-background"`,
		`for="app-toggle"`,
		`id="app-panel"`,
		`>Starter</span>`,
		`>Home</span>`,
		`>Users</span>`,
		`href="/login"`,
		`href="/register"`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Navbar() guest output missing %q in %s", check, got)
		}
	}
	if strings.Contains(got, `action="/logout"`) {
		t.Fatalf("Navbar() guest output unexpectedly rendered logout: %s", got)
	}
	if strings.Contains(got, `>Ada</span>`) {
		t.Fatalf("Navbar() guest output unexpectedly rendered user text: %s", got)
	}
}

func TestNavbarRendersAuthenticatedItems(t *testing.T) {
	got := testutil.RenderNode(t, testNavbar(true, "Ada", nil))

	checks := []string{
		`>Ada</span>`,
		`action="/logout"`,
		`value="csrf-token"`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Navbar() authenticated output missing %q in %s", check, got)
		}
	}
	if strings.Contains(got, `href="/login"`) {
		t.Fatalf("Navbar() authenticated output unexpectedly rendered login link: %s", got)
	}
}

func TestNavbarOmitsThemeItemWhenNodeIsNil(t *testing.T) {
	got := testutil.RenderNode(t, testNavbar(false, "", nil))

	if strings.Contains(got, `>Theme</span>`) {
		t.Fatalf("Navbar() unexpectedly rendered mobile theme row without node: %s", got)
	}
}

func TestNavbarRendersThemeItemWhenNodeIsProvided(t *testing.T) {
	theme := h.Button(h.Type("button"), g.Text("Theme toggle"))
	got := testutil.RenderNode(t, testNavbar(false, "", theme))

	if !strings.Contains(got, `Theme toggle`) {
		t.Fatalf("Navbar() output missing provided theme node: %s", got)
	}
	if !strings.Contains(got, `>Theme</span>`) {
		t.Fatalf("Navbar() output missing mobile theme label when node is provided: %s", got)
	}
}

func TestNavbarUsesDropdownAndAccordionGrouping(t *testing.T) {
	got := testutil.RenderNode(t, testNavbar(false, "", nil))

	checks := []string{
		`class="group relative"`,
		`class="absolute left-0 top-full z-50 mt-px flex min-w-[16rem] flex-col divide-y divide-border rounded-md border border-border bg-popover shadow-lg"`,
		`class="flex flex-col divide-y divide-border border-t border-border/70"`,
		`class="w-full divide-y divide-border rounded-lg border border-border"`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Navbar() grouping output missing %q in %s", check, got)
		}
	}
}

func TestNavbarMarksCurrentLink(t *testing.T) {
	got := testutil.RenderNode(t, Navbar(NavbarProps{
		ID:          "app",
		CurrentPath: "/users",
		Brand:       NavbarBrand{Label: "Starter", Href: "/"},
		Sections: []NavbarSection{{Items: []NavbarItem{
			NavLink{Label: "Users", Href: "/users"},
		}}},
	}))

	if !strings.Contains(got, `aria-current="page"`) {
		t.Fatalf("Navbar() output missing current-page marker: %s", got)
	}
}

func testNavbar(authenticated bool, userName string, theme g.Node) g.Node {
	var themeItem NavbarItem = NavNode{}
	if theme != nil {
		themeItem = NavNode{
			Desktop: theme,
			Mobile: h.Div(
				h.Class("flex items-center justify-between px-4 py-4 transition-colors hover:bg-accent/60"),
				h.Span(h.Class("text-sm text-muted-foreground"), g.Text("Theme")),
				theme,
			),
		}
	}

	return Navbar(NavbarProps{
		ID:              "app",
		IsAuthenticated: authenticated,
		Brand:           NavbarBrand{Label: "Starter", Href: "/"},
		Sections: []NavbarSection{
			{Items: []NavbarItem{
				NavLink{Label: "Home", Href: "/", Icon: componenticons.ChevronRight},
				NavGroup{Label: "Explore", Items: []NavbarItem{
					NavLink{Label: "Components", Href: "/_components"},
					NavGroup{Label: "Admin", Items: []NavbarItem{
						NavLink{Label: "Users", Href: "/users"},
					}},
				}},
			}},
			{Align: AlignEnd, Items: []NavbarItem{
				NavLink{Label: "Login", Href: "/login", Visibility: VisibilityGuest},
				NavLink{Label: "Register", Href: "/register", Visibility: VisibilityGuest},
				NavText{Text: userName, Visibility: VisibilityUser},
				NavForm{
					Label:      "Logout",
					Action:     "/logout",
					Visibility: VisibilityUser,
					HiddenFields: []NavHiddenField{{
						Name:  "_csrf",
						Value: "csrf-token",
					}},
				},
				themeItem,
			}},
		},
	})
}
