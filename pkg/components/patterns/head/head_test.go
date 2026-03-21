package head

import (
	"strings"
	"testing"

	"starter/pkg/components/testutil"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func TestHeadRendersSegmentedMetadataAssetsAndExtras(t *testing.T) {
	got := testutil.RenderNode(t, Head(Props{
		Meta: MetaProps{
			AppName:     "Starter",
			Title:       "Users",
			Description: "Manage users.",
			Keywords:    []string{"go", "echo"},
			FaviconHref: "/public/favicon.png",
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
		`<title>Starter | Users</title>`,
		`name="description" content="Manage users."`,
		`name="keywords" content="go, echo"`,
		`rel="icon" href="/public/favicon.png"`,
		`rel="stylesheet" href="/public/css/app.css"`,
		`name="csrf-token" content="csrf-token"`,
		`src="/public/js/app.js" defer`,
		`src="/public/js/htmx.min.js" defer`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered head missing %q:\n%s", want, got)
		}
	}
}

func TestMetatagsFallsBackToSingleTitleValue(t *testing.T) {
	tests := []struct {
		name string
		meta MetaProps
		want string
	}{
		{
			name: "app and page title",
			meta: MetaProps{AppName: "Starter", Title: "Home"},
			want: `<title>Starter | Home</title>`,
		},
		{
			name: "page title only",
			meta: MetaProps{Title: "Home"},
			want: `<title>Home</title>`,
		},
		{
			name: "app name only",
			meta: MetaProps{AppName: "Starter"},
			want: `<title>Starter</title>`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := testutil.RenderNode(t, Metatags(tc.meta))
			if !strings.Contains(got, tc.want) {
				t.Fatalf("rendered metatags missing %q:\n%s", tc.want, got)
			}
		})
	}
}
