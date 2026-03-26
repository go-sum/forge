package contactpartial_test

import (
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/go-sum/componentry/testutil"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/partial/contactpartial"
	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

var (
	contactRoutesOnce sync.Once
	contactRoutes     echo.Routes
)

func mustContactRoutes(t *testing.T) echo.Routes {
	t.Helper()
	contactRoutesOnce.Do(func() {
		e := echo.New()
		noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }
		route.Add(e, echo.Route{Method: http.MethodGet, Path: "/contact", Name: "contact.show", Handler: noOp})
		route.Add(e, echo.Route{Method: http.MethodPost, Path: "/contact", Name: "contact.submit", Handler: noOp})
		contactRoutes = e.Router().Routes()
	})
	return contactRoutes
}

func TestContactForm_showsFormFields(t *testing.T) {
	req := view.Request{
		CSRFToken:     "csrf-token",
		CSRFFieldName: "_csrf",
		Routes:        mustContactRoutes(t),
	}
	got := testutil.RenderNode(t, contactpartial.ContactForm(req, model.ContactFormData{}))

	wantSnippets := []string{
		`id="contact-form"`,
		`name="name"`,
		`name="email"`,
		`name="message"`,
		`Send message`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("ContactForm missing %q:\n%s", want, got)
		}
	}
}

func TestContactForm_showsSuccessStateWhenSent(t *testing.T) {
	req := view.Request{Routes: mustContactRoutes(t)}
	got := testutil.RenderNode(t, contactpartial.ContactForm(req, model.ContactFormData{Sent: true}))

	if !strings.Contains(got, "Message sent!") {
		t.Fatalf("expected success message in output:\n%s", got)
	}
	if strings.Contains(got, `name="name"`) {
		t.Error("form fields should not appear in success state")
	}
}

func TestContactForm_showsValidationErrors(t *testing.T) {
	req := view.Request{
		CSRFToken:     "csrf-token",
		CSRFFieldName: "_csrf",
		Routes:        mustContactRoutes(t),
	}
	data := model.ContactFormData{
		Values: model.ContactInput{Name: "Alice"},
		Errors: map[string][]string{
			"Email": {"Email is required."},
		},
	}
	got := testutil.RenderNode(t, contactpartial.ContactForm(req, data))

	if !strings.Contains(got, "Email is required.") {
		t.Fatalf("expected error message in output:\n%s", got)
	}
}
