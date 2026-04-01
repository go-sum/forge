package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sum/componentry/patterns/flash"
)

func TestHomeRendersFlashMessages(t *testing.T) {
	h := newTestHandler(fakeUserService{}, nil)
	c, rec := newRequestContext(http.MethodGet, "/", nil)
	setCSRFToken(c)
	setUserID(c, testUser.ID.String())

	flashRec := httptest.NewRecorder()
	if err := flash.Success(flashRec, "Saved"); err != nil {
		t.Fatalf("flash.Success() error = %v", err)
	}
	for _, cookie := range flashRec.Result().Cookies() {
		c.Request().AddCookie(cookie)
	}

	if err := h.Home(c); err != nil {
		t.Fatalf("Home() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Modern Web Starter") || !strings.Contains(body, "Saved") {
		t.Fatalf("body = %q", body)
	}
}

func TestComponentExamplesRenders(t *testing.T) {
	h := newTestHandler(fakeUserService{}, nil)
	c, rec := newRequestContext(http.MethodGet, "/_components", nil)
	setCSRFToken(c)

	if err := h.ComponentExamples(c); err != nil {
		t.Fatalf("ComponentExamples() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Component Examples") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
