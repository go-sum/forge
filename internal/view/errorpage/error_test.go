package errorpage

import (
	"strings"
	"testing"

	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/componentry/testutil"
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
		`Request ID: <code class="font-medium text-foreground">req-123</code>`,
		`<code class="font-medium text-foreground">req-123</code>`,
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
