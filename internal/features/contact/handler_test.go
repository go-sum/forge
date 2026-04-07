package contact

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/app/testutil"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/server/validate"
)

type fakeSubmitter struct {
	submitFn func(context.Context, model.ContactInput) error
}

func (f fakeSubmitter) Submit(ctx context.Context, input model.ContactInput) error {
	if f.submitFn != nil {
		return f.submitFn(ctx, input)
	}
	return errors.New("unexpected Submit call")
}

func newTestHandler(svc submitter) *Handler {
	return NewHandler(
		&config.Config{App: config.AppConfig{Security: config.SecurityConfig{CSRF: config.CSRFConfig{ContextKey: "csrf"}}}},
		svc,
		validate.New(),
	)
}

func TestFormRenders200(t *testing.T) {
	h := newTestHandler(fakeSubmitter{})
	c, rec := testutil.NewRequestContext(http.MethodGet, "/contact", nil)

	if err := h.Form(c); err != nil {
		t.Fatalf("Form returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "contact-form") {
		t.Errorf("expected contact-form in response body")
	}
}

func TestSubmitValidFormReturns200Fragment(t *testing.T) {
	called := false
	svc := fakeSubmitter{
		submitFn: func(_ context.Context, input model.ContactInput) error {
			called = true
			if input.Name != "Alice" {
				t.Errorf("expected Name=Alice, got %q", input.Name)
			}
			return nil
		},
	}
	h := newTestHandler(svc)

	form := url.Values{
		"name":    {"Alice"},
		"email":   {"alice@example.com"},
		"message": {"Hello there!"},
	}
	c, rec := testutil.NewFormContext(http.MethodPost, "/contact", form)
	testutil.SetCSRFToken(c)

	if err := h.Submit(c); err != nil {
		t.Fatalf("Submit returned unexpected error: %v", err)
	}
	if !called {
		t.Error("expected Submit to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestSubmitInvalidFormReturns422(t *testing.T) {
	h := newTestHandler(fakeSubmitter{
		submitFn: func(_ context.Context, _ model.ContactInput) error {
			t.Error("Submit should not be called on invalid form")
			return nil
		},
	})

	form := url.Values{"name": {"Alice"}}
	c, rec := testutil.NewFormContext(http.MethodPost, "/contact", form)
	testutil.SetCSRFToken(c)

	if err := h.Submit(c); err != nil {
		t.Fatalf("Submit returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestSubmitServiceErrorReturnsAppError(t *testing.T) {
	svc := fakeSubmitter{
		submitFn: func(_ context.Context, _ model.ContactInput) error {
			return errors.New("smtp down")
		},
	}
	h := newTestHandler(svc)

	form := url.Values{
		"name":    {"Bob"},
		"email":   {"bob@example.com"},
		"message": {"Test message"},
	}
	c, _ := testutil.NewFormContext(http.MethodPost, "/contact", form)
	testutil.SetCSRFToken(c)

	err := h.Submit(c)
	testutil.AssertAppErrorStatus(t, err, http.StatusServiceUnavailable)
}
