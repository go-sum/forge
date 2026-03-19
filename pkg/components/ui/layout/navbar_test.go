package layout

import (
	"strings"
	"testing"

	"starter/config"
)

func TestNavMenuRendersConfiguredGuestMenu(t *testing.T) {
	got := renderNode(t, NavMenu(NavMenuProps{
		ID:        "app",
		Config:    testNavConfig(),
		CSRFToken: "csrf-token",
	}))

	checks := []string{
		`class="w-full border-b bg-background"`,
		`for="app-toggle"`,
		`id="app-panel"`,
		`>Starter</span>`,
		`>Home</a>`,
		`>Features</span>`,
		`>Users</a>`,
		`href="/login"`,
		`href="/register"`,
		`data-theme-toggle=""`,
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

func TestNavMenuRendersConfiguredAuthenticatedMenu(t *testing.T) {
	got := renderNode(t, NavMenu(NavMenuProps{
		ID:              "app",
		Config:          testNavConfig(),
		CSRFToken:       "csrf-token",
		IsAuthenticated: true,
		UserName:        "Ada",
	}))

	checks := []string{
		`>Ada</span>`,
		`action="/logout"`,
		`value="csrf-token"`,
		`data-theme-toggle=""`,
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

func testNavConfig() config.NavConfig {
	return config.NavConfig{
		Brand: config.NavBrand{Label: "Starter", Href: "/"},
		Sections: []config.NavSection{
			{Items: []config.NavItem{
				{Label: "Home", Href: "/"},
				{Label: "Features", Items: []config.NavItem{
					{Label: "Admin", Items: []config.NavItem{
						{Label: "Users", Href: "/users"},
					}},
				}},
			}},
			{Items: []config.NavItem{
				{Label: "Login", Href: "/login", Visibility: "guest"},
				{Label: "Register", Href: "/register", Visibility: "guest"},
				{Type: "user_name", Visibility: "user"},
				{Type: "logout", Label: "Logout", Visibility: "user"},
				{Type: "theme_toggle"},
			}},
		},
	}
}

func TestNavMenuUsesAccordionStyleForNestedDesktopAndMobileMenus(t *testing.T) {
	got := renderNode(t, NavMenu(NavMenuProps{
		ID:     "app",
		Config: testNavConfig(),
	}))

	checks := []string{
		`class="absolute z-50 min-w-[16rem] border border-border bg-popover shadow-lg left-0 top-full mt-px flex flex-col divide-y divide-border"`,
		`class="border-t border-border/70 bg-background"`,
		`class="w-full divide-y divide-border border-y border-border"`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("NavMenu() accordion output missing %q in %s", check, got)
		}
	}
	if strings.Contains(got, `left-full top-0 ml-2`) {
		t.Fatalf("NavMenu() unexpectedly rendered desktop flyout submenu: %s", got)
	}
}
