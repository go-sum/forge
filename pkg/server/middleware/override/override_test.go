package override_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sum/server/middleware/override"
	"github.com/labstack/echo/v5"
)

// newPOSTContext builds an Echo context for a POST request with an
// application/x-www-form-urlencoded body containing the given form field.
func newPOSTContext(e *echo.Echo, field, value string) (*echo.Context, *httptest.ResponseRecorder) {
	var body string
	if field != "" && value != "" {
		body = field + "=" + value
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// newMethodContext builds an Echo context for a request with an explicit method
// and no body.
func newMethodContext(e *echo.Echo, method string) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// captureMethod is a handler that records the effective request method.
func captureMethod(captured *string) echo.HandlerFunc {
	return func(c *echo.Context) error {
		*captured = c.Request().Method
		return c.NoContent(http.StatusOK)
	}
}

func TestPOSTWithDeleteOverride(t *testing.T) {
	e := echo.New()
	mw := override.Middleware()

	var got string
	c, _ := newPOSTContext(e, "_method", "DELETE")
	if err := mw(captureMethod(&got))(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != http.MethodDelete {
		t.Errorf("method = %q, want %q", got, http.MethodDelete)
	}
}

func TestPOSTWithPutOverride(t *testing.T) {
	e := echo.New()
	mw := override.Middleware()

	var got string
	c, _ := newPOSTContext(e, "_method", "PUT")
	if err := mw(captureMethod(&got))(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != http.MethodPut {
		t.Errorf("method = %q, want %q", got, http.MethodPut)
	}
}

func TestPOSTWithPatchOverride(t *testing.T) {
	e := echo.New()
	mw := override.Middleware()

	var got string
	c, _ := newPOSTContext(e, "_method", "PATCH")
	if err := mw(captureMethod(&got))(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != http.MethodPatch {
		t.Errorf("method = %q, want %q", got, http.MethodPatch)
	}
}

func TestPOSTWithLowercaseDelete(t *testing.T) {
	e := echo.New()
	mw := override.Middleware()

	var got string
	c, _ := newPOSTContext(e, "_method", "delete")
	if err := mw(captureMethod(&got))(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != http.MethodDelete {
		t.Errorf("method = %q, want %q (case insensitive)", got, http.MethodDelete)
	}
}

func TestPOSTWithNoOverrideField(t *testing.T) {
	e := echo.New()
	mw := override.Middleware()

	nextCalled := false
	var got string
	next := func(c *echo.Context) error {
		nextCalled = true
		got = c.Request().Method
		return c.NoContent(http.StatusOK)
	}

	c, _ := newPOSTContext(e, "", "")
	if err := mw(next)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Error("next was not called when no override field is present")
	}
	if got != http.MethodPost {
		t.Errorf("method = %q, want %q", got, http.MethodPost)
	}
}

func TestGETIgnored(t *testing.T) {
	e := echo.New()
	mw := override.Middleware()

	var got string
	c, _ := newMethodContext(e, http.MethodGet)
	if err := mw(captureMethod(&got))(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != http.MethodGet {
		t.Errorf("method = %q, want %q (GET must not be overridden)", got, http.MethodGet)
	}
}

func TestPOSTWithGetOverrideIsViolation(t *testing.T) {
	e := echo.New()
	mw := override.Middleware()

	nextCalled := false
	next := func(c *echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	}

	c, _ := newPOSTContext(e, "_method", "GET")
	err := mw(next)(c)
	if err == nil {
		t.Fatal("expected violation error, got nil")
	}
	if nextCalled {
		t.Error("next was called despite invalid override value")
	}
	type statuser interface{ StatusCode() int }
	s, ok := err.(statuser)
	if !ok {
		t.Fatalf("error does not implement StatusCode(): %T", err)
	}
	if s.StatusCode() != http.StatusBadRequest {
		t.Errorf("StatusCode() = %d, want %d", s.StatusCode(), http.StatusBadRequest)
	}
}

func TestPOSTWithHeadOverrideIsViolation(t *testing.T) {
	e := echo.New()
	mw := override.Middleware()

	c, _ := newPOSTContext(e, "_method", "HEAD")
	err := mw(func(c *echo.Context) error { return nil })(c)
	if err == nil {
		t.Fatal("expected violation for HEAD override, got nil")
	}
	type statuser interface{ StatusCode() int }
	s, ok := err.(statuser)
	if !ok {
		t.Fatalf("error does not implement StatusCode(): %T", err)
	}
	if s.StatusCode() != http.StatusBadRequest {
		t.Errorf("StatusCode() = %d, want %d", s.StatusCode(), http.StatusBadRequest)
	}
}

func TestPOSTWithArbitraryVerbIsViolation(t *testing.T) {
	e := echo.New()
	mw := override.Middleware()

	c, _ := newPOSTContext(e, "_method", "FOOBAR")
	err := mw(func(c *echo.Context) error { return nil })(c)
	if err == nil {
		t.Fatal("expected violation for arbitrary verb, got nil")
	}
	type statuser interface{ StatusCode() int }
	s, ok := err.(statuser)
	if !ok {
		t.Fatalf("error does not implement StatusCode(): %T", err)
	}
	if s.StatusCode() != http.StatusBadRequest {
		t.Errorf("StatusCode() = %d, want %d", s.StatusCode(), http.StatusBadRequest)
	}
}

func TestCustomFieldName(t *testing.T) {
	e := echo.New()
	mw := override.NewWithConfig(override.Config{
		FormField: "_m",
	})

	var got string
	c, _ := newPOSTContext(e, "_m", "DELETE")
	if err := mw(captureMethod(&got))(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != http.MethodDelete {
		t.Errorf("method = %q, want %q", got, http.MethodDelete)
	}
}

func TestDefaultFieldNotReadForCustomFieldName(t *testing.T) {
	e := echo.New()
	mw := override.NewWithConfig(override.Config{
		FormField: "_m",
	})

	// Sends "_method=DELETE" (default field name) but middleware reads "_m"
	var got string
	c, _ := newPOSTContext(e, "_method", "DELETE")
	if err := mw(captureMethod(&got))(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Method should stay POST because "_m" field is absent
	if got != http.MethodPost {
		t.Errorf("method = %q, want %q (custom field not found)", got, http.MethodPost)
	}
}

func TestSkipperBypasses(t *testing.T) {
	e := echo.New()
	mw := override.NewWithConfig(override.Config{
		Skipper: func(*echo.Context) bool { return true },
	})

	var got string
	c, _ := newPOSTContext(e, "_method", "DELETE")
	if err := mw(captureMethod(&got))(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Method should stay POST because skipper returned true
	if got != http.MethodPost {
		t.Errorf("method = %q, want %q (skipped)", got, http.MethodPost)
	}
}
