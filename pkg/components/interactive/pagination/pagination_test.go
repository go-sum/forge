package pagination

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
	testutil "starter/pkg/components/testutil"
)

func TestPaginationLinkAndNavStates(t *testing.T) {
	active := testutil.RenderNode(t, Link("/users?page=2", true, g.Text("2")))
	if !strings.Contains(active, ` aria-current="page"`) || !strings.Contains(active, ` href="/users?page=2"`) {
		t.Fatalf("Link() output = %s", active)
	}

	previous := testutil.RenderNode(t, Previous("/users?page=1", true))
	if strings.Contains(previous, `<a`) || !strings.Contains(previous, `<span`) ||
		!strings.Contains(previous, `aria-disabled="true"`) || !strings.Contains(previous, ` aria-label="Go to previous page"`) {
		t.Fatalf("Previous() output = %s", previous)
	}

	next := testutil.RenderNode(t, Next("/users?page=3", false))
	if !strings.Contains(next, ` href="/users?page=3"`) || !strings.Contains(next, `>Next</span>`) {
		t.Fatalf("Next() output = %s", next)
	}
	if !strings.Contains(next, `focus-visible:ring-[3px]`) || !strings.Contains(previous, `focus-visible:ring-[3px]`) {
		t.Fatalf("Previous()/Next() output missing focus-visible styling: prev=%s next=%s", previous, next)
	}

	withAttrs := testutil.RenderNode(t, Previous("/users?page=1", false, g.Attr("hx-get", "/users?page=1"), g.Attr("hx-target", "#users-list-region")))
	if !strings.Contains(withAttrs, `hx-get="/users?page=1"`) || !strings.Contains(withAttrs, `hx-target="#users-list-region"`) {
		t.Fatalf("Previous() output missing extra attrs: %s", withAttrs)
	}
}

func TestPaginationRootRendersNavigationWrapper(t *testing.T) {
	got := testutil.RenderNode(t, Root(Content(Item(Ellipsis()))))
	if !strings.Contains(got, ` aria-label="pagination"`) || !strings.Contains(got, `aria-hidden="true">…</span>`) {
		t.Fatalf("Root() output = %s", got)
	}
}
