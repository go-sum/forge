package ratelimit

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

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
	err := Middleware(Config{Store: NewMemoryStore(100)})(func(c *echo.Context) error {
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
// The default policy normalization gives Burst=1 when both Rate and Burst are 0.
// Two requests from the same IP therefore exhaust the single-token burst and
// trigger the deny path on the second call.
func TestMiddlewareDeniedRequestReturnsTypedError(t *testing.T) {
	e := echo.New()
	mw := Middleware(Config{Store: NewMemoryStoreWithConfig(MemoryStoreConfig{Rate: 0, Burst: 0})})

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
		Store: NewMemoryStore(100),
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

type failingStore struct{}

func (failingStore) Allow(_ string) (bool, error) {
	return false, errors.New("backend unavailable")
}

// When the store returns (false, err), the DenyHandler is invoked with the error.
// This follows the reference implementation pattern: a failing store is treated
// as a denial (429) rather than an internal error (500).
func TestMiddlewareReturnsTypedErrorWhenStoreFails(t *testing.T) {
	_, c := newContext(http.MethodGet, "/")

	err := Middleware(Config{
		Store: failingStore{},
	})(func(c *echo.Context) error { return nil })(c)

	if err == nil {
		t.Fatal("Middleware() error = nil, want typed denial error")
	}

	type statusCoder interface{ StatusCode() int }
	sc, ok := err.(statusCoder)
	if !ok {
		t.Fatalf("error %T does not implement StatusCode()", err)
	}
	if sc.StatusCode() != http.StatusTooManyRequests {
		t.Fatalf("StatusCode() = %d, want %d", sc.StatusCode(), http.StatusTooManyRequests)
	}
}

// When the Skipper returns true the middleware must call next even with an
// exhausted store.
func TestMiddlewareSkipperBypassesRateLimit(t *testing.T) {
	_, c := newContext(http.MethodGet, "/")

	// Zero-rate store: every Allow() returns false.
	store := NewMemoryStoreWithConfig(MemoryStoreConfig{Rate: 0, Burst: 0})
	// Exhaust the single burst token first via a direct call.
	store.Allow("192.0.2.1") //nolint:errcheck

	var nextCalled bool
	err := Middleware(Config{
		Store:   store,
		Skipper: func(*echo.Context) bool { return true },
	})(func(c *echo.Context) error {
		nextCalled = true
		return nil
	})(c)

	if err != nil {
		t.Fatalf("Skipper=true returned error: %v", err)
	}
	if !nextCalled {
		t.Fatal("next was not called when Skipper returned true")
	}
}

// A custom DenyHandler must be invoked when the store denies a request.
func TestMiddlewareCustomDenyHandlerIsInvoked(t *testing.T) {
	e := echo.New()
	store := NewMemoryStoreWithConfig(MemoryStoreConfig{Rate: 0, Burst: 0})

	var denyHandlerCalled bool
	mw := Middleware(Config{
		Store: store,
		DenyHandler: func(c *echo.Context, identifier string, err error) error {
			denyHandlerCalled = true
			return &violation{code: http.StatusTooManyRequests, msg: denyMessage}
		},
	})

	noop := func(c *echo.Context) error { return nil }

	// First request consumes the single burst token.
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := mw(noop)(e.NewContext(req1, httptest.NewRecorder())); err != nil {
		t.Fatalf("first request: unexpected error: %v", err)
	}

	// Second request should trigger DenyHandler.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	_ = mw(noop)(e.NewContext(req2, httptest.NewRecorder()))

	if !denyHandlerCalled {
		t.Fatal("custom DenyHandler was not called on rate-limit deny")
	}
}

// A custom ErrorHandler must be invoked when IdentifierExtractor returns an error.
func TestMiddlewareCustomErrorHandlerIsInvoked(t *testing.T) {
	_, c := newContext(http.MethodGet, "/")

	var errorHandlerCalled bool
	err := Middleware(Config{
		Store: NewMemoryStore(100),
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			return "", errors.New("extractor failure")
		},
		ErrorHandler: func(c *echo.Context, err error) error {
			errorHandlerCalled = true
			return &violation{code: http.StatusInternalServerError, msg: "Service temporarily unavailable"}
		},
	})(func(c *echo.Context) error { return nil })(c)

	if err == nil {
		t.Fatal("expected error from ErrorHandler, got nil")
	}
	if !errorHandlerCalled {
		t.Fatal("custom ErrorHandler was not called when extractor returned error")
	}
}

// Config{}.ToMiddleware() must return a non-nil error when Store is nil.
func TestToMiddlewareErrorsWhenStoreIsNil(t *testing.T) {
	_, err := Config{}.ToMiddleware()
	if err == nil {
		t.Fatal("ToMiddleware() with nil Store should return error, got nil")
	}
}

// After the ExpiresIn window elapses, stale visitor entries must be swept so
// that a new request from the same IP gets a fresh limiter with a full burst.
func TestMemoryStoreExpiresStaleVisitors(t *testing.T) {
	store := NewMemoryStoreWithConfig(MemoryStoreConfig{
		Rate:      0,
		Burst:     0, // normalizes to burst=1
		ExpiresIn: 50 * time.Millisecond,
	})

	// Exhaust the single burst token.
	allowed, _ := store.Allow("192.0.2.1")
	if !allowed {
		t.Fatal("first Allow() should be allowed (fresh limiter)")
	}
	denied, _ := store.Allow("192.0.2.1")
	if denied {
		t.Fatal("second Allow() should be denied (burst exhausted)")
	}

	// Wait for the expiry window to pass.
	time.Sleep(60 * time.Millisecond)

	// A subsequent Allow() triggers cleanup and installs a new limiter.
	// The visitor's lastSeen is stale, so after cleanup a new visitor is created.
	// We need to trigger cleanup; call Allow() with a different IP to force sweep.
	store.Allow("192.0.2.2") //nolint:errcheck

	// Now the original IP should get a fresh limiter after cleanup.
	refreshed, _ := store.Allow("192.0.2.1")
	if !refreshed {
		t.Fatal("after expiry, Allow() should be allowed again (fresh limiter after cleanup)")
	}
}

// Concurrent calls to Allow() must not cause data races or panics.
func TestMemoryStoreConcurrentAccess(t *testing.T) {
	store := NewMemoryStore(100)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			store.Allow("shared-key") //nolint:errcheck
		}()
	}
	wg.Wait()
}
