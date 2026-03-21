package page

import (
	"strings"
	"testing"

	"starter/internal/view"
	"starter/pkg/components/patterns/flash"
	"starter/pkg/components/testutil"
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
		`Welcome`,
		`A Go starter with Echo, HTMX, and Tailwind.`,
		`Saved`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered home page missing %q:\n%s", want, got)
		}
	}
}
