package pagination

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

func TestPaginationLinkAndNavStates(t *testing.T) {
	active := renderNode(t, Link("/users?page=2", true, g.Text("2")))
	if !strings.Contains(active, ` aria-current="page"`) || !strings.Contains(active, ` href="/users?page=2"`) {
		t.Fatalf("Link() output = %s", active)
	}

	previous := renderNode(t, Previous("/users?page=1", true))
	if strings.Contains(previous, `<a href=`) || !strings.Contains(previous, `<a class="`) || !strings.Contains(previous, ` aria-label="Go to previous page"`) {
		t.Fatalf("Previous() output = %s", previous)
	}

	next := renderNode(t, Next("/users?page=3", false))
	if !strings.Contains(next, ` href="/users?page=3"`) || !strings.Contains(next, `>Next</span>`) {
		t.Fatalf("Next() output = %s", next)
	}
}

func TestPaginationRootRendersNavigationWrapper(t *testing.T) {
	got := renderNode(t, Root(Content(Item(Ellipsis()))))
	if !strings.Contains(got, ` aria-label="pagination"`) || !strings.Contains(got, `aria-hidden="true">…</span>`) {
		t.Fatalf("Root() output = %s", got)
	}
}
