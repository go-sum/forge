package head

import (
	"strings"
	"testing"

	"github.com/go-sum/componentry/testutil"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func TestHeadRendersSegmentedMetadataAssetsAndExtras(t *testing.T) {
	got := testutil.RenderNode(t, Head(Props{
		Meta: MetaProps{
			Title:       "Users",
			Description: "Manage users.",
			Keywords:    []string{"go", "echo"},
			FaviconHref: "/public/favicon.png",
			OGImage:     "https://example.com/og.png",
		},
		Stylesheets: []Stylesheet{{Href: "/public/css/app.css"}},
		Extra: []g.Node{
			h.Meta(h.Name("csrf-token"), h.Content("csrf-token")),
		},
		Scripts: []Script{
			{Src: "/public/js/app.js", Defer: true},
			{Src: "/public/js/htmx.min.js", Defer: true},
		},
	}))

	wantSnippets := []string{
		`<head>`,
		`<title>Users</title>`,
		`name="description" content="Manage users."`,
		`name="keywords" content="go, echo"`,
		`rel="icon" href="/public/favicon.png"`,
		`rel="stylesheet" href="/public/css/app.css"`,
		`name="csrf-token" content="csrf-token"`,
		`src="/public/js/app.js" defer`,
		`src="/public/js/htmx.min.js" defer`,
		`property="og:title" content="Users"`,
		`property="og:type" content="website"`,
		`property="og:description" content="Manage users."`,
		`property="og:image" content="https://example.com/og.png"`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered head missing %q:\n%s", want, got)
		}
	}
}

