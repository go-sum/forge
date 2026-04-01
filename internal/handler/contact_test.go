package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/server/validate"

	"github.com/labstack/echo/v5"
)

type fakeContactService struct {
	submitFn func(context.Context, model.ContactInput) error
}

func (f fakeContactService) Submit(ctx context.Context, input model.ContactInput) error {
	if f.submitFn != nil {
		return f.submitFn(ctx, input)
	}
	return errors.New("unexpected Submit call")
}

func newContactHandler(svc contactService) *Handler {
	e := echo.New()
	registerTestRoutes(e)
	return &Handler{
		services:  handlerServices{Contact: svc},
		validator: validate.New(),
		cfg:       &config.Config{App: config.AppConfig{Keys: testKeys}},
		routes:    func() echo.Routes { return e.Router().Routes() },
	}
}

func TestContactForm_renders200(t *testing.T) {
	h := newContactHandler(fakeContactService{})
	c, rec := newRequestContext(http.MethodGet, "/contact", nil)

	if err := h.ContactForm(c); err != nil {
		t.Fatalf("ContactForm returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "contact-form") {
		t.Errorf("expected contact-form in response body")
	}
}

func TestContactSubmit_validForm_returns200Fragment(t *testing.T) {
	called := false
	svc := fakeContactService{
		submitFn: func(_ context.Context, input model.ContactInput) error {
			called = true
			if input.Name != "Alice" {
				t.Errorf("expected Name=Alice, got %q", input.Name)
			}
			return nil
		},
	}
	h := newContactHandler(svc)

	form := url.Values{
		"name":    {"Alice"},
		"email":   {"alice@example.com"},
		"message": {"Hello there!"},
	}
	c, rec := newFormContext(http.MethodPost, "/contact", form)
	setCSRFToken(c)

	if err := h.ContactSubmit(c); err != nil {
		t.Fatalf("ContactSubmit returned unexpected error: %v", err)
	}
	if !called {
		t.Error("expected Submit to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestContactSubmit_invalidForm_returns422(t *testing.T) {
	h := newContactHandler(fakeContactService{
		submitFn: func(_ context.Context, _ model.ContactInput) error {
			t.Error("Submit should not be called on invalid form")
			return nil
		},
	})

	// Missing required fields
	form := url.Values{"name": {"Alice"}}
	c, rec := newFormContext(http.MethodPost, "/contact", form)
	setCSRFToken(c)

	if err := h.ContactSubmit(c); err != nil {
		t.Fatalf("ContactSubmit returned unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestContactSubmit_serviceError_returnsAppError(t *testing.T) {
	svc := fakeContactService{
		submitFn: func(_ context.Context, _ model.ContactInput) error {
			return errors.New("smtp down")
		},
	}
	h := newContactHandler(svc)

	form := url.Values{
		"name":    {"Bob"},
		"email":   {"bob@example.com"},
		"message": {"Test message"},
	}
	c, _ := newFormContext(http.MethodPost, "/contact", form)
	setCSRFToken(c)

	err := h.ContactSubmit(c)
	assertAppErrorStatus(t, err, http.StatusServiceUnavailable)
}
