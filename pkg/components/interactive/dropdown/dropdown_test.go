package dropdown

import (
	"strings"
	"testing"

	testutil "starter/pkg/components/testutil"
	g "maragu.dev/gomponents"
)

func TestTriggerRendersStyledSummary(t *testing.T) {
	got := testutil.RenderNode(t, Trigger(TriggerProps{}, g.Text("Options")))

	if !strings.Contains(got, `<summary`) {
		t.Fatalf("Trigger() output missing summary element: %s", got)
	}
	if strings.Contains(got, `<button`) {
		t.Fatalf("Trigger() output unexpectedly nested a button: %s", got)
	}
	if !strings.Contains(got, `Options`) {
		t.Fatalf("Trigger() output missing label text: %s", got)
	}
}

func TestDisabledLinkItemOmitsHref(t *testing.T) {
	got := testutil.RenderNode(t, Item(ItemProps{Label: "View Profile", Href: "/profile", Disabled: true}))

	if strings.Contains(got, ` href="/profile"`) {
		t.Fatalf("Item() output kept href for disabled link: %s", got)
	}
	if !strings.Contains(got, ` aria-disabled="true"`) {
		t.Fatalf("Item() output missing aria-disabled for disabled link: %s", got)
	}
	if !strings.Contains(got, ` role="menuitem"`) {
		t.Fatalf("Item() output missing menuitem role: %s", got)
	}
}
