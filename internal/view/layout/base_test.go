package layout

import (
	"strings"
	"testing"

	"github.com/go-sum/componentry/patterns/flash"
	"github.com/go-sum/componentry/testutil"

	g "maragu.dev/gomponents"
)

func TestPageInjectsCSRFAssetsAndFlash(t *testing.T) {
	got := testutil.RenderNode(t, Page(Props{
		Title:         "Home",
		FaviconPath:   "/public/img/favicon.ico",
		CSRFFieldName: "_csrf",
		CSRFToken:     "csrf-token",
		Flash: []flash.Message{{
			Type: flash.TypeSuccess,
			Text: "Saved",
		}},
		Children: []g.Node{g.Text("Body content")},
	}))

	wantSnippets := []string{
		`<title>Home</title>`,
		`rel="icon" href="/public/img/favicon.ico"`,
		`name="csrf-token" content="csrf-token"`,
		`hx-headers="{&#34;X-CSRF-Token&#34;:&#34;csrf-token&#34;}"`,
		`id="toast-container"`,
		`Saved`,
		`Body content`,
		`src="/public/js/app.js"`,
		`src="/public/js/htmx.min.js"`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered page missing %q:\n%s", want, got)
		}
	}
}
