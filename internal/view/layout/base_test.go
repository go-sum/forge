package layout

import (
	"strings"
	"testing"

	"github.com/go-sum/componentry/patterns/flash"
	"github.com/go-sum/componentry/testutil"
	"github.com/go-sum/forge/config"

	g "maragu.dev/gomponents"
)

func TestPageInjectsCSRFAssetsAndFlash(t *testing.T) {
	got := testutil.RenderNode(t, Page(Props{
		Title:          "Home",
		FaviconPath:    "/public/img/favicon.ico",
		CSRFFieldName:  "_csrf",
		CSRFHeaderName: "X-CSRF-Token",
		CSRFToken:      "csrf-token",
		Flash: []flash.Message{{
			Type: flash.TypeSuccess,
			Text: "Saved",
		}},
		Children: []g.Node{g.Text("Body content")},
	}))

	wantSnippets := []string{
		`<title>Home</title>`,
		`rel="icon" href="/public/img/favicon.ico"`,
		`name="csrf-token" content="csrf-token"`,
		`hx-headers="{&#34;X-CSRF-Token&#34;:&#34;csrf-token&#34;}"`,
		`id="toast-container"`,
		`Saved`,
		`Body content`,
		`src="/public/js/app.js"`,
		`src="/public/js/htmx.min.js"`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered page missing %q:\n%s", want, got)
		}
	}
}

func TestPageRendersAccountDropdownSlotsWithMenuRowStyling(t *testing.T) {
	got := testutil.RenderNode(t, Page(Props{
		CSRFFieldName:   "_csrf",
		CSRFHeaderName:  "X-CSRF-Token",
		CSRFToken:       "csrf-token",
		SignoutPath:      "/profile/signout",
		IsAuthenticated: true,
		UserName:        "John",
		NavConfig: config.NavConfig{
			Sections: []config.NavSection{{
				Align: "end",
				Items: []config.NavItem{{
					Label: "Account",
					Items: []config.NavItem{
						{Slot: "user_name", Visibility: "user"},
						{Slot: "signout", Visibility: "user"},
					},
				}},
			}},
		},
	}))

	wantSnippets := []string{
		`class="absolute left-0 top-full z-50 mt-px flex min-w-[16rem] flex-col divide-y divide-border rounded-md border border-border bg-popover shadow-lg"`,
		`<span class="block w-full px-4 py-3 text-sm font-medium text-foreground"><span>John</span></span>`,
		`<button type="submit" class="block px-4 py-3 text-sm font-medium w-full text-left outline-none transition-colors hover:bg-accent/60 hover:text-accent-foreground focus-visible:ring-[3px] focus-visible:ring-ring/50"><span>Signout</span></button>`,
		`<input type="hidden" name="_csrf" value="csrf-token">`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered page missing %q:\n%s", want, got)
		}
	}
}
