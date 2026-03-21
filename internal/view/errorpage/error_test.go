package errorpage

import (
	"strings"
	"testing"

	"starter/pkg/components/testutil"
)

func TestPageRendersDebugDetailAndFallbacks(t *testing.T) {
	got := testutil.RenderNode(t, Page(Props{
		Status:          500,
		Message:         "Something went wrong.",
		RequestID:       "req-123",
		Debug:           true,
		TechnicalDetail: "database timeout",
		CSRFToken:       "csrf-token",
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
