package dropdown

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
	testutil "github.com/y-goweb/componentry/testutil"
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
	if !strings.Contains(got, `focus-visible:ring-[3px]`) {
		t.Fatalf("Trigger() output missing focus-visible styling: %s", got)
	}
}

func TestDisabledLinkItemOmitsHrefAndDoesNotClaimMenuSemantics(t *testing.T) {
	got := testutil.RenderNode(t, Item(ItemProps{Label: "View Profile", Href: "/profile", Disabled: true}))

	if strings.Contains(got, ` href="/profile"`) {
		t.Fatalf("Item() output kept href for disabled link: %s", got)
	}
	if !strings.Contains(got, ` aria-disabled="true"`) {
		t.Fatalf("Item() output missing aria-disabled for disabled link: %s", got)
	}
	if strings.Contains(got, ` role="menuitem"`) {
		t.Fatalf("Item() output unexpectedly claimed menuitem role: %s", got)
	}
	if !strings.Contains(got, `focus-visible:ring-[3px]`) {
		t.Fatalf("Item() output missing focus-visible styling: %s", got)
	}
}

func TestContentUsesPopoverSemanticsInsteadOfMenuRole(t *testing.T) {
	got := testutil.RenderNode(t, Content(g.Text("Body")))

	if strings.Contains(got, ` role="menu"`) {
		t.Fatalf("Content() output unexpectedly claimed menu role: %s", got)
	}
}
