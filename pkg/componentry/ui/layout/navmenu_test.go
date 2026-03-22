package layout

import (
	"strings"
	"testing"

	testutil "github.com/go-sum/componentry/testutil"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func TestNavMenuRendersConfiguredGuestMenu(t *testing.T) {
	got := testutil.RenderNode(t, NavMenu(NavMenuProps{
		ID:     "app",
		Config: testNavMenuConfig(),
		Slots: NavSlots{
			"theme_toggle": ControlSlot("Theme", h.Button(h.Type("button"), g.Text("Theme toggle"))),
		},
	}))

	checks := []string{
		`class="w-full border-b bg-background"`,
		`for="app-toggle"`,
		`id="app-panel"`,
		`>Starter</span>`,
		`href="/login"`,
		`Theme toggle`,
		`>Theme</span>`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("NavMenu() guest output missing %q in %s", check, got)
		}
	}
	if strings.Contains(got, `action="/logout"`) {
		t.Fatalf("NavMenu() guest output unexpectedly rendered logout: %s", got)
	}
}

func TestNavMenuRendersAuthenticatedMenuWithSlots(t *testing.T) {
	got := testutil.RenderNode(t, NavMenu(NavMenuProps{
		ID:              "app",
		Config:          testNavMenuConfig(),
		IsAuthenticated: true,
		Slots: NavSlots{
			"user_name": TextSlot("Ada"),
			"logout": FormSlot(FormSlotProps{
				Label:  "Logout",
				Action: "/logout",
				HiddenFields: []NavHiddenField{{
					Name:  "_csrf",
					Value: "csrf-token",
				}},
			}),
		},
	}))

	checks := []string{
		`>Ada</span>`,
		`action="/logout"`,
		`value="csrf-token"`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("NavMenu() authenticated output missing %q in %s", check, got)
		}
	}
	if strings.Contains(got, `href="/login"`) {
		t.Fatalf("NavMenu() authenticated output unexpectedly rendered login link: %s", got)
	}
}

func TestNavMenuOmitsMissingSlots(t *testing.T) {
	got := testutil.RenderNode(t, NavMenu(NavMenuProps{
		ID:     "app",
		Config: testNavMenuConfig(),
	}))

	if strings.Contains(got, `>Theme</span>`) {
		t.Fatalf("NavMenu() unexpectedly rendered theme slot without a registered node: %s", got)
	}
}

func TestNavMenuMarksCurrentLinkFromCurrentPath(t *testing.T) {
	got := testutil.RenderNode(t, NavMenu(NavMenuProps{
		ID:          "app",
		Config:      testNavMenuConfig(),
		CurrentPath: "/users",
	}))

	if !strings.Contains(got, `aria-current="page"`) {
		t.Fatalf("NavMenu() output missing current-page marker: %s", got)
	}
}

func testNavMenuConfig() NavConfig {
	return NavConfig{
		Brand: NavbarBrand{Label: "Starter", Href: "/"},
		Sections: []NavSection{
			{Items: []NavItem{
				{Label: "Home", Href: "/"},
				{Label: "Explore", Items: []NavItem{
					{Label: "Components", Href: "/_components"},
					{Label: "Admin", Items: []NavItem{{Label: "Users", Href: "/users"}}},
				}},
			}},
			{Align: AlignEnd, Items: []NavItem{
				{Label: "Login", Href: "/login", Visibility: VisibilityGuest},
				{Label: "Register", Href: "/register", Visibility: VisibilityGuest},
				{Slot: "user_name", Visibility: VisibilityUser},
				{Slot: "logout", Visibility: VisibilityUser},
				{Slot: "theme_toggle"},
			}},
		},
	}
}
