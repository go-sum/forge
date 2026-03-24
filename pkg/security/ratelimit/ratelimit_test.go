package ratelimit

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

// A request within the rate limit must pass through to the next handler.
func TestMiddlewareAllowsRequestWithinLimit(t *testing.T) {
	_, c := newContext(http.MethodGet, "/")

	var called bool
	err := Middleware(Config{Rate: 100, Burst: 200})(func(c *echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})(c)

	if err != nil {
		t.Fatalf("Middleware() within limit returned error: %v", err)
	}
	if !called {
		t.Fatal("next handler was not called within limit")
	}
}

// When the rate limit is exceeded the middleware must return a typed error
// with StatusCode() == 429 and a non-empty PublicMessage().
//
// Echo's memory store defaults Burst to max(1, ceil(Rate)) when Burst is 0,
// so Rate=0 gives one free token. Two requests from the same IP exhaust the
// single-token burst and trigger the DenyHandler on the second call.
func TestMiddlewareDeniedRequestReturnsTypedError(t *testing.T) {
	e := echo.New()
	mw := Middleware(Config{Rate: 0, Burst: 0}) // Burst defaults to 1 internally

	noop := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }

	// First request — consumes the single burst token (passes through).
	req1 := httptest.NewRequest(http.MethodPost, "/signin", nil)
	c1 := e.NewContext(req1, httptest.NewRecorder())
	if err := mw(noop)(c1); err != nil {
		t.Fatalf("first request returned unexpected error: %v", err)
	}

	// Second request from the same IP — bucket empty, rate = 0, DenyHandler fires.
	req2 := httptest.NewRequest(http.MethodPost, "/signin", nil)
	c2 := e.NewContext(req2, httptest.NewRecorder())
	err := mw(func(c *echo.Context) error {
		t.Fatal("next handler must not be called when rate limit is exceeded")
		return nil
	})(c2)

	if err == nil {
		t.Fatal("second request within exhausted limit returned nil error")
	}

	type statusCoder interface{ StatusCode() int }
	type publicMessager interface{ PublicMessage() string }

	sc, ok := err.(statusCoder)
	if !ok {
		t.Fatalf("error %T does not implement StatusCode()", err)
	}
	if sc.StatusCode() != http.StatusTooManyRequests {
		t.Fatalf("StatusCode() = %d, want %d", sc.StatusCode(), http.StatusTooManyRequests)
	}

	pm, ok := err.(publicMessager)
	if !ok {
		t.Fatalf("error %T does not implement PublicMessage()", err)
	}
	if pm.PublicMessage() == "" {
		t.Fatal("PublicMessage() must not be empty")
	}
}

// A custom IdentifierExtractor must be used when provided.
func TestMiddlewareUsesCustomIdentifierExtractor(t *testing.T) {
	_, c := newContext(http.MethodGet, "/")

	var extractorCalled bool
	mw := Middleware(Config{
		Rate:  100,
		Burst: 200,
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			extractorCalled = true
			return "custom-key", nil
		},
	})

	_ = mw(func(c *echo.Context) error { return nil })(c)

	if !extractorCalled {
		t.Fatal("custom IdentifierExtractor was not called")
	}
}
