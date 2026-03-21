package errorpage

import (
	"strings"
	"testing"

	"starter/internal/view"
	"starter/pkg/components/testutil"
)

func TestPageRendersDebugDetailAndFallbacks(t *testing.T) {
	got := testutil.RenderNode(t, Page(view.Request{
		CSRFToken: "csrf-token",
	}, Props{
		Status:          500,
		Message:         "Something went wrong.",
		RequestID:       "req-123",
		Debug:           true,
		TechnicalDetail: "database timeout",
	}))

	wantSnippets := []string{
		`Internal Server Error`,
		`HTTP 500`,
		`Request ID: req-123`,
		`Technical Detail`,
		`database timeout`,
		`href="/"`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered page missing %q:\n%s", want, got)
		}
	}
}
