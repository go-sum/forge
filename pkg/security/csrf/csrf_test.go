package csrf

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
		ContextKey:   "csrf",
		HeaderName:   "X-CSRF-Token",
		FormField:    "_csrf",
		CookieName:   "_csrf",
		CookieSecure: false,
	}
}

// GET should always pass through and store a token in the Echo context.
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

// GET should store a token value in the Echo context.
// For browsers that send Sec-Fetch-Site: same-origin Echo stores its sentinel
// constant; for others it stores the generated cookie value. Either way the
// context value must be a non-empty string.
func TestMiddlewareStoresTokenInContext(t *testing.T) {
	const ctxKey = "csrf"
	_, c := newContext(http.MethodGet, "/")

	err := Middleware(testConfig())(func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)

	if err != nil {
		t.Fatalf("Middleware() error = %v", err)
	}
	token, ok := c.Get(ctxKey).(string)
	if !ok || token == "" {
		t.Fatalf("context key %q: got %q, want non-empty string", ctxKey, token)
	}
}

// POST without any CSRF cookie or token should return a typed violation error
// with StatusCode() == 403 and a non-empty PublicMessage().
func TestMiddlewareUnsafeMethodWithoutTokenReturnsTypedError(t *testing.T) {
	_, c := newContext(http.MethodPost, "/users")

	err := Middleware(testConfig())(func(c *echo.Context) error {
		t.Fatal("next handler must not be called when CSRF fails")
		return nil
	})(c)

	if err == nil {
		t.Fatal("Middleware() on POST without token returned nil error")
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

// Cross-site request (Sec-Fetch-Site: cross-site) on a POST should also
// return a typed 403 violation (not a raw *echo.HTTPError).
func TestMiddlewareCrossSiteRequestReturnsTypedError(t *testing.T) {
	_, c := newContext(http.MethodPost, "/users")
	c.Request().Header.Set("Sec-Fetch-Site", "cross-site")

	err := Middleware(testConfig())(func(c *echo.Context) error {
		t.Fatal("next handler must not be called for cross-site POST")
		return nil
	})(c)

	if err == nil {
		t.Fatal("Middleware() on cross-site POST returned nil error")
	}

	type statusCoder interface{ StatusCode() int }
	sc, ok := err.(statusCoder)
	if !ok {
		t.Fatalf("error %T does not implement StatusCode()", err)
	}
	if sc.StatusCode() != http.StatusForbidden {
		t.Fatalf("StatusCode() = %d, want %d", sc.StatusCode(), http.StatusForbidden)
	}
}
