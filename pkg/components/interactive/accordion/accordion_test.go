package accordion

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

func TestAccordionRendersNativeStructure(t *testing.T) {
	got := renderNode(t, Root(Item(Trigger(g.Text("Account")), Content(g.Text("Manage your account settings.")))))

	checks := []string{
		`<div class="w-full divide-y divide-border rounded-lg border">`,
		`<details class="px-4">`,
		`<summary class="flex w-full items-center justify-between py-4 text-sm font-medium transition-all hover:underline text-left cursor-pointer">Account`,
		`<div class="pb-4 text-sm text-muted-foreground">Manage your account settings.</div>`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("accordion output missing %q in %s", check, got)
		}
	}
}
