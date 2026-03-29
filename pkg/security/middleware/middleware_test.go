package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-sum/security/csrf"
	"github.com/go-sum/security/fetchmeta"
	"github.com/go-sum/security/origin"
	"github.com/go-sum/security/token"
	"github.com/labstack/echo/v5"
)

func TestCrossOriginGuardAllowsVerifiedRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := CrossOriginGuard(
		origin.Policy{
			Enabled:         true,
			CanonicalOrigin: "https://example.com",
			RequireHeader:   true,
		},
		fetchmeta.Policy{
			Enabled:                 true,
			AllowedSites:            []string{"same-origin", "same-site"},
			AllowedModes:            []string{"cors", "navigate", "same-origin"},
			FallbackWhenMissing:     true,
			RejectCrossSiteNavigate: true,
		},
	)(func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err != nil {
		t.Fatalf("CrossOriginGuard() error = %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCrossOriginGuardReturnsTypedError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := CrossOriginGuard(
		origin.Policy{
			Enabled:         true,
			CanonicalOrigin: "https://example.com",
			RequireHeader:   true,
		},
		fetchmeta.Policy{
			Enabled:             true,
			FallbackWhenMissing: true,
		},
	)(func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err == nil {
		t.Fatal("CrossOriginGuard() error = nil")
	}

	mwErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("err = %T, want *Error", err)
	}
	if mwErr.Status != http.StatusForbidden {
		t.Fatalf("status = %d", mwErr.Status)
	}
	if mwErr.PublicMessage() == "" {
		t.Fatal("PublicMessage() should not be empty")
	}
}

func TestToMiddlewareRejectsNoPolicies(t *testing.T) {
	_, err := Config{}.ToMiddleware()
	if err == nil {
		t.Fatal("ToMiddleware() error = nil, want error for no enabled policies")
	}
	want := "crossorigin: at least one of OriginPolicy or FetchPolicy must be enabled"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func TestMiddlewareSkipperSkipsCheck(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	// No origin headers — would fail without skipper.
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	mw := Middleware(Config{
		Skipper: func(*echo.Context) bool { return true },
		OriginPolicy: origin.Policy{
			Enabled:         true,
			CanonicalOrigin: "https://example.com",
			RequireHeader:   true,
		},
	})
	err := mw(func(c *echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err != nil {
		t.Fatalf("Middleware() error = %v", err)
	}
	if !called {
		t.Fatal("next handler was not called")
	}
}

func TestMiddlewareSkipperDoesNotSkipSafeMethod(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	mw := Middleware(Config{
		Skipper: func(*echo.Context) bool { return false },
		OriginPolicy: origin.Policy{
			Enabled:         true,
			CanonicalOrigin: "https://example.com",
			RequireHeader:   true,
		},
	})
	err := mw(func(c *echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err != nil {
		t.Fatalf("Middleware() error = %v", err)
	}
	if !called {
		t.Fatal("next handler was not called for safe method")
	}
}

func TestConfigToMiddlewareAllowsVerifiedRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw, err := Config{
		OriginPolicy: origin.Policy{
			Enabled:         true,
			CanonicalOrigin: "https://example.com",
			RequireHeader:   true,
		},
		FetchPolicy: fetchmeta.Policy{
			Enabled:                 true,
			AllowedSites:            []string{"same-origin", "same-site"},
			AllowedModes:            []string{"cors", "navigate", "same-origin"},
			FallbackWhenMissing:     true,
			RejectCrossSiteNavigate: true,
		},
	}.ToMiddleware()
	if err != nil {
		t.Fatalf("ToMiddleware() error = %v", err)
	}

	handlerErr := mw(func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})(c)
	if handlerErr != nil {
		t.Fatalf("middleware error = %v", handlerErr)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestConfigToMiddlewareReturnsTypedError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw, err := Config{
		OriginPolicy: origin.Policy{
			Enabled:         true,
			CanonicalOrigin: "https://example.com",
			RequireHeader:   true,
		},
		FetchPolicy: fetchmeta.Policy{
			Enabled:             true,
			FallbackWhenMissing: true,
		},
	}.ToMiddleware()
	if err != nil {
		t.Fatalf("ToMiddleware() error = %v", err)
	}

	handlerErr := mw(func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})(c)
	if handlerErr == nil {
		t.Fatal("middleware error = nil, want *Error")
	}

	mwErr, ok := handlerErr.(*Error)
	if !ok {
		t.Fatalf("err = %T, want *Error", handlerErr)
	}
	if mwErr.Status != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", mwErr.Status, http.StatusForbidden)
	}
	if mwErr.PublicMessage() == "" {
		t.Fatal("PublicMessage() should not be empty")
	}
}

// ---------------------------------------------------------------------------
// Composition tests: CSRF + CrossOriginGuard chained on the same request
// ---------------------------------------------------------------------------
//
// These tests prove the intended layering:
//   - CrossOriginGuard validates request context (Origin + Fetch Metadata headers)
//   - CSRF validates client capability (HMAC-signed token)
//   - Both run in sequence; CrossOriginGuard first, CSRF second
//
// This is the recommended composition for browser-facing mutation endpoints.
// API endpoints that use bearer tokens or CORS may omit CSRF.
// ---------------------------------------------------------------------------

var compositionKey = []byte("composition-test-key-32-bytes!!X")

func compositionOriginPolicy() origin.Policy {
	return origin.Policy{
		Enabled:         true,
		CanonicalOrigin: "https://example.com",
		RequireHeader:   true,
	}
}

func compositionFetchPolicy() fetchmeta.Policy {
	return fetchmeta.Policy{
		Enabled:                 true,
		AllowedSites:            []string{"same-origin", "same-site"},
		AllowedModes:            []string{"cors", "navigate", "same-origin"},
		FallbackWhenMissing:     true,
		RejectCrossSiteNavigate: true,
	}
}

// TestCompositionBothPassWhenRequestIsValid verifies that a request with valid
// origin headers AND a valid CSRF token passes through both middleware.
func TestCompositionBothPassWhenRequestIsValid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")

	tok, err := token.Issue(compositionKey, "csrf", time.Hour)
	if err != nil {
		t.Fatalf("token.Issue() error = %v", err)
	}
	req.Header.Set("X-CSRF-Token", tok)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	guard := Middleware(Config{
		OriginPolicy: compositionOriginPolicy(),
		FetchPolicy:  compositionFetchPolicy(),
	})
	csrfMW := csrf.Middleware(csrf.Config{
		Key:        compositionKey,
		ContextKey: "csrf",
		HeaderName: "X-CSRF-Token",
		FormField:  "_csrf",
	})

	called := false
	// Chain: CrossOriginGuard wraps CSRF wraps handler
	handler := guard(csrfMW(func(c *echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	}))

	if err := handler(c); err != nil {
		t.Fatalf("chained middleware error = %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

// TestCompositionGuardRejectsPriorToCSRF verifies that CrossOriginGuard blocks
// a cross-origin request before CSRF ever runs.
func TestCompositionGuardRejectsPriorToCSRF(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	// Missing Origin and Fetch Metadata headers — CrossOriginGuard should reject.
	// Include a valid CSRF token to prove CSRF never runs.
	tok, _ := token.Issue(compositionKey, "csrf", time.Hour)
	req.Header.Set("X-CSRF-Token", tok)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	guard := Middleware(Config{
		OriginPolicy: compositionOriginPolicy(),
		FetchPolicy:  compositionFetchPolicy(),
	})
	csrfMW := csrf.Middleware(csrf.Config{
		Key:        compositionKey,
		ContextKey: "csrf",
		HeaderName: "X-CSRF-Token",
		FormField:  "_csrf",
	})

	handler := guard(csrfMW(func(c *echo.Context) error {
		t.Fatal("handler must not be called when CrossOriginGuard rejects")
		return nil
	}))

	err := handler(c)
	if err == nil {
		t.Fatal("expected CrossOriginGuard rejection, got nil")
	}

	mwErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("err = %T, want *Error (CrossOriginGuard typed error)", err)
	}
	if mwErr.Status != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", mwErr.Status, http.StatusForbidden)
	}
}

// TestCompositionCSRFRejectsAfterGuardPasses verifies that when CrossOriginGuard
// passes but the CSRF token is missing, the CSRF middleware blocks the request.
func TestCompositionCSRFRejectsAfterGuardPasses(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	// No CSRF token — should be rejected by CSRF middleware.

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	guard := Middleware(Config{
		OriginPolicy: compositionOriginPolicy(),
		FetchPolicy:  compositionFetchPolicy(),
	})
	csrfMW := csrf.Middleware(csrf.Config{
		Key:        compositionKey,
		ContextKey: "csrf",
		HeaderName: "X-CSRF-Token",
		FormField:  "_csrf",
	})

	handler := guard(csrfMW(func(c *echo.Context) error {
		t.Fatal("handler must not be called when CSRF token is missing")
		return nil
	}))

	err := handler(c)
	if err == nil {
		t.Fatal("expected CSRF rejection, got nil")
	}

	// CSRF violation implements StatusCode() int — not *Error.
	type statusCoder interface{ StatusCode() int }
	sc, ok := err.(statusCoder)
	if !ok {
		t.Fatalf("err = %T, does not implement StatusCode()", err)
	}
	if sc.StatusCode() != http.StatusForbidden {
		t.Fatalf("StatusCode() = %d, want %d", sc.StatusCode(), http.StatusForbidden)
	}
}
