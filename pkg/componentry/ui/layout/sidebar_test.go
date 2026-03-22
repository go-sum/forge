package layout

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
	testutil "github.com/y-goweb/componentry/testutil"
)

func TestSidebarUsesInstanceScopedIDs(t *testing.T) {
	got := testutil.RenderNode(t, Sidebar(SidebarProps{ID: "admin", Nav: g.Text("nav")}))

	if !strings.Contains(got, ` id="admin-backdrop"`) {
		t.Fatalf("Sidebar() output missing backdrop id: %s", got)
	}
	if !strings.Contains(got, ` id="admin-panel"`) {
		t.Fatalf("Sidebar() output missing panel id: %s", got)
	}
	if !strings.Contains(got, ` id="admin-toggle"`) {
		t.Fatalf("Sidebar() output missing toggle input id: %s", got)
	}
	if strings.Contains(got, ` role="dialog"`) || strings.Contains(got, ` aria-modal="true"`) {
		t.Fatalf("Sidebar() output unexpectedly claimed dialog semantics: %s", got)
	}
	if !strings.Contains(got, ` aria-label="Mobile navigation"`) {
		t.Fatalf("Sidebar() output missing mobile navigation label: %s", got)
	}
}

func TestToggleAttrsReferenceScopedSidebar(t *testing.T) {
	got := testutil.RenderNode(t, g.El("label", ToggleAttrs("admin")...))

	if !strings.Contains(got, ` for="admin-toggle"`) {
		t.Fatalf("ToggleAttrs() output missing scoped label target: %s", got)
	}
	if !strings.Contains(got, ` aria-controls="admin-panel"`) {
		t.Fatalf("ToggleAttrs() output missing scoped aria-controls: %s", got)
	}
}
