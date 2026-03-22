package redirect

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuilderUsesHXRedirectForHTMXRequests(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	if err := New(rec, req).To("/login").Go(); err != nil {
		t.Fatalf("Go() error = %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("HX-Redirect"); got != "/login" {
		t.Fatalf("HX-Redirect = %q", got)
	}
}

func TestBuilderFallsBackToHTTPRedirect(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	if err := New(rec, req).To("/users?page=2").StatusCode(http.StatusFound).Go(); err != nil {
		t.Fatalf("Go() error = %v", err)
	}
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "/users?page=2" {
		t.Fatalf("Location = %q", got)
	}
}

func TestBuilderBoostedRequestUsesHTTPRedirect(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Boosted", "true")
	rec := httptest.NewRecorder()

	if err := New(rec, req).To("/dashboard").Go(); err != nil {
		t.Fatalf("Go() error = %v", err)
	}
	// Boosted requests behave like full-page navigations — use standard redirect.
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/dashboard" {
		t.Fatalf("Location = %q", got)
	}
}
