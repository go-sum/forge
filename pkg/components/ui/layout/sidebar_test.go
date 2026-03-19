package layout

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

func TestSidebarUsesInstanceScopedIDs(t *testing.T) {
	got := renderNode(t, Sidebar(SidebarProps{ID: "admin", Nav: g.Text("nav")}))

	if !strings.Contains(got, ` id="admin-backdrop"`) {
		t.Fatalf("Sidebar() output missing backdrop id: %s", got)
	}
	if !strings.Contains(got, ` id="admin-panel"`) {
		t.Fatalf("Sidebar() output missing panel id: %s", got)
	}
	if !strings.Contains(got, ` id="admin-toggle"`) {
		t.Fatalf("Sidebar() output missing toggle input id: %s", got)
	}
}

func TestToggleAttrsReferenceScopedSidebar(t *testing.T) {
	got := renderNode(t, g.El("label", ToggleAttrs("admin")...))

	if !strings.Contains(got, ` for="admin-toggle"`) {
		t.Fatalf("ToggleAttrs() output missing scoped label target: %s", got)
	}
	if !strings.Contains(got, ` aria-controls="admin-panel"`) {
		t.Fatalf("ToggleAttrs() output missing scoped aria-controls: %s", got)
	}
}
