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
		`href="/signin"`,
		`Theme toggle`,
		`>Theme</span>`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("NavMenu() guest output missing %q in %s", check, got)
		}
	}
	if strings.Contains(got, `action="/signout"`) {
		t.Fatalf("NavMenu() guest output unexpectedly rendered signout: %s", got)
	}
}

func TestNavMenuRendersAuthenticatedMenuWithSlots(t *testing.T) {
	got := testutil.RenderNode(t, NavMenu(NavMenuProps{
		ID:              "app",
		Config:          testNavMenuConfig(),
		IsAuthenticated: true,
		Slots: NavSlots{
			"user_name": TextSlot("Ada"),
			"signout": FormSlot(FormSlotProps{
				Label:  "Signout",
				Action: "/signout",
				HiddenFields: []NavHiddenField{{
					Name:  "_csrf",
					Value: "csrf-token",
				}},
			}),
		},
	}))

	checks := []string{
		`>Ada</span>`,
		`action="/signout"`,
		`value="csrf-token"`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("NavMenu() authenticated output missing %q in %s", check, got)
		}
	}
	if strings.Contains(got, `href="/signin"`) {
		t.Fatalf("NavMenu() authenticated output unexpectedly rendered signin link: %s", got)
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
				{Label: "Signin", Href: "/signin", Visibility: VisibilityGuest},
				{Label: "Signup", Href: "/signup", Visibility: VisibilityGuest},
				{Slot: "user_name", Visibility: VisibilityUser},
				{Slot: "signout", Visibility: VisibilityUser},
				{Slot: "theme_toggle"},
			}},
		},
	}
}
