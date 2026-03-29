package csrf

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-sum/security/token"
	"github.com/labstack/echo/v5"
)

func newContext(method, path string) (*echo.Echo, *echo.Context) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e, e.NewContext(req, rec)
}

func testConfig() Config {
	return Config{
		Key:        []byte("test-signing-key-32-bytes-padded!"),
		TokenTTL:   3600,
		ContextKey: "csrf",
		HeaderName: "X-CSRF-Token",
		FormField:  "_csrf",
	}
}

// GET should always pass through and store an HMAC token in the Echo context.
func TestMiddlewareSafeMethodPassesThrough(t *testing.T) {
	_, c := newContext(http.MethodGet, "/")

	var handlerCalled bool
	err := Middleware(testConfig())(func(c *echo.Context) error {
		handlerCalled = true
		return c.NoContent(http.StatusOK)
	})(c)

	if err != nil {
		t.Fatalf("Middleware() on GET returned error: %v", err)
	}
	if !handlerCalled {
		t.Fatal("next handler was not called on GET")
	}
}

// GET should store a non-empty HMAC token string in the Echo context.
func TestMiddlewareStoresTokenInContext(t *testing.T) {
	const ctxKey = "csrf"
	_, c := newContext(http.MethodGet, "/")

	err := Middleware(testConfig())(func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)

	if err != nil {
		t.Fatalf("Middleware() error = %v", err)
	}
	tok, ok := c.Get(ctxKey).(string)
	if !ok || tok == "" {
		t.Fatalf("context key %q: got %q, want non-empty HMAC token", ctxKey, tok)
	}
}

// The context token must be a verifiable HMAC token (not a random cookie value).
func TestMiddlewareContextTokenIsVerifiable(t *testing.T) {
	cfg := testConfig()
	_, c := newContext(http.MethodGet, "/")

	_ = Middleware(cfg)(func(c *echo.Context) error { return nil })(c)

	tok, _ := c.Get(cfg.ContextKey).(string)
	if err := token.Verify(cfg.Key, scope, tok); err != nil {
		t.Fatalf("token.Verify() on context token = %v, want nil", err)
	}
}

func TestMiddlewareUsesDefaultTTLWhenUnset(t *testing.T) {
	cfg := testConfig()
	cfg.TokenTTL = 0

	_, c := newContext(http.MethodGet, "/")
	err := Middleware(cfg)(func(c *echo.Context) error { return nil })(c)
	if err != nil {
		t.Fatalf("Middleware() error = %v", err)
	}

	tok, ok := c.Get(cfg.ContextKey).(string)
	if !ok || tok == "" {
		t.Fatalf("context key %q: got %q, want non-empty HMAC token", cfg.ContextKey, tok)
	}
	if err := token.Verify(cfg.Key, scope, tok); err != nil {
		t.Fatalf("token.Verify() on default-ttl token = %v, want nil", err)
	}
}

// POST without any CSRF token should return a typed violation error (403).
func TestMiddlewareUnsafeMethodWithoutTokenReturnsTypedError(t *testing.T) {
	_, c := newContext(http.MethodPost, "/users")

	err := Middleware(testConfig())(func(c *echo.Context) error {
		t.Fatal("next handler must not be called when CSRF fails")
		return nil
	})(c)

	assertViolation(t, err)
}

// POST with a valid token in the header should succeed.
func TestMiddlewareValidHeaderTokenPasses(t *testing.T) {
	cfg := testConfig()
	tok, _ := token.Issue(cfg.Key, scope, time.Hour)

	_, c := newContext(http.MethodPost, "/users")
	c.Request().Header.Set(cfg.HeaderName, tok)

	var called bool
	err := Middleware(cfg)(func(c *echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})(c)

	if err != nil {
		t.Fatalf("Middleware() with valid header token = %v, want nil", err)
	}
	if !called {
		t.Fatal("next handler was not called")
	}
}

// POST with a valid token in the form field should succeed.
func TestMiddlewareValidFormTokenPasses(t *testing.T) {
	cfg := testConfig()
	tok, _ := token.Issue(cfg.Key, scope, time.Hour)

	body := url.Values{cfg.FormField: {tok}}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	var called bool
	err := Middleware(cfg)(func(c *echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})(c)

	if err != nil {
		t.Fatalf("Middleware() with valid form token = %v, want nil", err)
	}
	if !called {
		t.Fatal("next handler was not called")
	}
}

// POST with a tampered token should return a typed 403 violation.
func TestMiddlewareTamperedTokenReturnsViolation(t *testing.T) {
	_, c := newContext(http.MethodPost, "/users")
	c.Request().Header.Set("X-CSRF-Token", "tampered-not-a-real-token")

	err := Middleware(testConfig())(func(c *echo.Context) error {
		t.Fatal("next handler must not be called for tampered token")
		return nil
	})(c)

	assertViolation(t, err)
}

// POST with an expired token should return a typed 403 violation.
func TestMiddlewareExpiredTokenReturnsViolation(t *testing.T) {
	cfg := testConfig()
	expired, _ := token.Issue(cfg.Key, scope, -time.Second)

	_, c := newContext(http.MethodPost, "/users")
	c.Request().Header.Set(cfg.HeaderName, expired)

	err := Middleware(cfg)(func(c *echo.Context) error {
		t.Fatal("next handler must not be called for expired token")
		return nil
	})(c)

	assertViolation(t, err)
}

// POST with a token signed by a different key should return a typed 403 violation.
func TestMiddlewareWrongKeyReturnsViolation(t *testing.T) {
	wrongKey := []byte("different-signing-key-32-bytes!!")
	tok, _ := token.Issue(wrongKey, scope, time.Hour)

	_, c := newContext(http.MethodPost, "/users")
	c.Request().Header.Set("X-CSRF-Token", tok)

	err := Middleware(testConfig())(func(c *echo.Context) error {
		t.Fatal("next handler must not be called for wrong-key token")
		return nil
	})(c)

	assertViolation(t, err)
}

// ToMiddleware must reject a key shorter than 32 bytes.
func TestToMiddlewareRejectsShortKey(t *testing.T) {
	cfg := testConfig()
	cfg.Key = []byte("too-short")

	_, err := cfg.ToMiddleware()
	if err == nil {
		t.Fatal("ToMiddleware() with short key: expected error, got nil")
	}
	if err.Error() != "csrf: Key must be at least 32 bytes" {
		t.Fatalf("ToMiddleware() error = %q, want %q", err.Error(), "csrf: Key must be at least 32 bytes")
	}
}

// ToMiddleware must default ContextKey to "csrf" when empty.
func TestToMiddlewareDefaultsEmptyContextKey(t *testing.T) {
	cfg := testConfig()
	cfg.ContextKey = ""

	mw, err := cfg.ToMiddleware()
	if err != nil {
		t.Fatalf("ToMiddleware() error = %v", err)
	}

	_, c := newContext(http.MethodGet, "/")
	if err := mw(func(c *echo.Context) error { return nil })(c); err != nil {
		t.Fatalf("middleware error = %v", err)
	}

	tok, ok := c.Get("csrf").(string)
	if !ok || tok == "" {
		t.Fatal("expected non-empty token stored under default key \"csrf\"")
	}
}

// ToMiddleware must default HeaderName to "X-CSRF-Token" when empty.
func TestToMiddlewareDefaultsEmptyHeaderName(t *testing.T) {
	cfg := testConfig()
	cfg.HeaderName = ""

	mw, err := cfg.ToMiddleware()
	if err != nil {
		t.Fatalf("ToMiddleware() error = %v", err)
	}

	tok, _ := token.Issue(cfg.Key, scope, time.Hour)
	_, c := newContext(http.MethodPost, "/")
	c.Request().Header.Set("X-CSRF-Token", tok)

	var called bool
	if err := mw(func(c *echo.Context) error { called = true; return nil })(c); err != nil {
		t.Fatalf("middleware error = %v", err)
	}
	if !called {
		t.Fatal("next handler was not called")
	}
}

// ToMiddleware must default FormField to "_csrf" when empty.
func TestToMiddlewareDefaultsEmptyFormField(t *testing.T) {
	cfg := testConfig()
	cfg.FormField = ""

	mw, err := cfg.ToMiddleware()
	if err != nil {
		t.Fatalf("ToMiddleware() error = %v", err)
	}

	tok, _ := token.Issue(cfg.Key, scope, time.Hour)
	body := url.Values{"_csrf": {tok}}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	var called bool
	if err := mw(func(c *echo.Context) error { called = true; return nil })(c); err != nil {
		t.Fatalf("middleware error = %v", err)
	}
	if !called {
		t.Fatal("next handler was not called")
	}
}

// Skipper returning true must bypass CSRF validation entirely.
func TestMiddlewareSkipperSkipsCheck(t *testing.T) {
	cfg := testConfig()
	cfg.Skipper = func(*echo.Context) bool { return true }

	_, c := newContext(http.MethodPost, "/users")
	// No token set — should still pass because skipper returns true.

	var called bool
	err := Middleware(cfg)(func(c *echo.Context) error {
		called = true
		return nil
	})(c)

	if err != nil {
		t.Fatalf("Middleware() with skipper=true returned error: %v", err)
	}
	if !called {
		t.Fatal("next handler was not called when skipper returned true")
	}
}

// Skipper returning false must enforce CSRF validation.
func TestMiddlewareSkipperDoesNotSkipWhenFalse(t *testing.T) {
	cfg := testConfig()
	cfg.Skipper = func(*echo.Context) bool { return false }

	_, c := newContext(http.MethodPost, "/users")
	// No token — should fail.

	err := Middleware(cfg)(func(c *echo.Context) error {
		t.Fatal("next handler must not be called when CSRF fails")
		return nil
	})(c)

	assertViolation(t, err)
}

// Custom TokenExtractor must be used instead of header/form lookup.
func TestMiddlewareCustomTokenExtractor(t *testing.T) {
	cfg := testConfig()
	tok, _ := token.Issue(cfg.Key, scope, time.Hour)
	cfg.TokenExtractor = func(c *echo.Context) (string, error) {
		return tok, nil
	}

	_, c := newContext(http.MethodPost, "/users")
	// No header or form token — extractor provides it.

	var called bool
	err := Middleware(cfg)(func(c *echo.Context) error {
		called = true
		return nil
	})(c)

	if err != nil {
		t.Fatalf("Middleware() with custom extractor returned error: %v", err)
	}
	if !called {
		t.Fatal("next handler was not called with custom extractor")
	}
}

// assertViolation checks that err is a non-nil *violation with StatusCode 403
// and a non-empty PublicMessage.
func assertViolation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected a violation error, got nil")
	}
	type statusCoder interface{ StatusCode() int }
	type publicMessager interface{ PublicMessage() string }

	sc, ok := err.(statusCoder)
	if !ok {
		t.Fatalf("error %T does not implement StatusCode()", err)
	}
	if sc.StatusCode() != http.StatusForbidden {
		t.Fatalf("StatusCode() = %d, want %d", sc.StatusCode(), http.StatusForbidden)
	}
	pm, ok := err.(publicMessager)
	if !ok {
		t.Fatalf("error %T does not implement PublicMessage()", err)
	}
	if pm.PublicMessage() == "" {
		t.Fatal("PublicMessage() must not be empty")
	}
}
