package breadcrumb

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
)

func renderNode(t *testing.T, node g.Node) string {
	t.Helper()

	var buf strings.Builder
	if err := node.Render(&buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return buf.String()
}

func TestBreadcrumbRendersAccessibleTrail(t *testing.T) {
	got := renderNode(t, Root(List(
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
