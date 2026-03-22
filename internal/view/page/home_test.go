package page

import (
	"strings"
	"testing"

	"github.com/y-goweb/foundry/internal/view"
	"github.com/y-goweb/componentry/patterns/flash"
	"github.com/y-goweb/componentry/testutil"
)

func TestHomePageRendersWelcomeAndFlash(t *testing.T) {
	got := testutil.RenderNode(t, HomePage(view.Request{
		CSRFToken: "csrf-token",
		Flash: []flash.Message{{
			Type: flash.TypeSuccess,
			Text: "Saved",
		}},
	}))

	wantSnippets := []string{
		`Modern Web Starter`,
		`Build server-rendered apps without giving up interaction quality.`,
		`Browse Components`,
		`Sign In`,
		`Saved`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered home page missing %q:\n%s", want, got)
		}
	}
}
