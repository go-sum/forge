package page

import (
	"strings"
	"testing"

	"github.com/go-sum/componentry/testutil"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/view"
)

func TestContactPage_rendersFormAndHeading(t *testing.T) {
	got := testutil.RenderNode(t, ContactPage(view.Request{
		CSRFToken:      "csrf-token",
		CSRFHeaderName: "X-CSRF-Token",
		Routes:         mustPageRoutes(t),
	}, model.ContactFormData{}))

	wantSnippets := []string{
		"Contact us",
		`id="contact-form"`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("ContactPage missing %q:\n%s", want, got)
		}
	}
}
