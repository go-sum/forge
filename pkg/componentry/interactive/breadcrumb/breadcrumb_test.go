package breadcrumb

import (
	"strings"
	"testing"

	testutil "github.com/go-sum/componentry/testutil"
	g "maragu.dev/gomponents"
)

func TestBreadcrumbRendersAccessibleTrail(t *testing.T) {
	got := testutil.RenderNode(t, Root(List(
		Item(Link("/", g.Text("Home"))),
		Item(Separator()),
		Item(Page(g.Text("Settings"))),
	)))

	checks := []string{
		`<nav aria-label="breadcrumb" class="flex">`,
		`<ol class="flex items-center flex-wrap gap-1 text-sm">`,
		`<a href="/" class="text-muted-foreground hover:text-foreground hover:underline flex items-center gap-1.5 transition-colors">Home</a>`,
		`<span class="mx-2 text-muted-foreground" aria-hidden="true">/</span>`,
		`<span class="font-medium text-foreground flex items-center gap-1.5" aria-current="page">Settings</span>`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("breadcrumb output missing %q in %s", check, got)
		}
	}
}
