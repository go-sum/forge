package cors_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/security/cors"
	"github.com/labstack/echo/v5"
)

func newContext(e *echo.Echo, method, origin string) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, "/", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func okHandler(c *echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

// TestExactOriginAllowed verifies that an exact-match allowed origin sets the
// Access-Control-Allow-Origin header to that origin.
func TestExactOriginAllowed(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode:         cors.OriginModeExact,
		AllowOrigins: []string{"https://example.com"},
	})

	c, rec := newContext(e, http.MethodGet, "https://example.com")
	err := mw(okHandler)(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got != "https://example.com" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "https://example.com")
	}
}

// TestExactOriginDenied verifies that a disallowed origin does NOT set the
// ACAO header, but the next handler is still called (browser enforces CORS).
func TestExactOriginDenied(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode:         cors.OriginModeExact,
		AllowOrigins: []string{"https://example.com"},
	})

	nextCalled := false
	next := func(c *echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "ok")
	}

	c, rec := newContext(e, http.MethodGet, "https://evil.com")
	err := mw(next)(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Error("next handler was not called for disallowed origin")
	}
	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

// TestExactWildcard verifies that "*" as an AllowOrigins entry reflects "*"
// regardless of the request origin.
func TestExactWildcard(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode:         cors.OriginModeExact,
		AllowOrigins: []string{"*"},
	})

	c, rec := newContext(e, http.MethodGet, "https://anything.com")
	if err := mw(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
}

// TestExactEmptyOriginsError verifies that an empty AllowOrigins list causes
// ToMiddleware to return an error.
func TestExactEmptyOriginsError(t *testing.T) {
	cfg := cors.Config{
		Mode:         cors.OriginModeExact,
		AllowOrigins: []string{},
	}
	_, err := cfg.ToMiddleware()
	if err == nil {
		t.Fatal("expected error for empty AllowOrigins, got nil")
	}
}

// TestRegexOriginMatches verifies that a matching regexp pattern sets ACAO.
func TestRegexOriginMatches(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode:         cors.OriginModeRegex,
		RegexOrigins: []string{`^https://.*\.example\.com$`},
	})

	c, rec := newContext(e, http.MethodGet, "https://sub.example.com")
	if err := mw(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got != "https://sub.example.com" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "https://sub.example.com")
	}
}

// TestRegexOriginNoMatch verifies that a non-matching origin does NOT set ACAO
// but still calls next.
func TestRegexOriginNoMatch(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode:         cors.OriginModeRegex,
		RegexOrigins: []string{`^https://.*\.example\.com$`},
	})

	nextCalled := false
	next := func(c *echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "ok")
	}

	c, rec := newContext(e, http.MethodGet, "https://evil.com")
	if err := mw(next)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Error("next handler was not called for non-matching regex origin")
	}
	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

// TestRegexInvalidPatternError verifies that an unparseable regexp causes
// ToMiddleware to return an error.
func TestRegexInvalidPatternError(t *testing.T) {
	cfg := cors.Config{
		Mode:         cors.OriginModeRegex,
		RegexOrigins: []string{`[invalid`},
	}
	_, err := cfg.ToMiddleware()
	if err == nil {
		t.Fatal("expected error for invalid regexp pattern, got nil")
	}
}

// TestRegexEmptyPatternsError verifies that empty RegexOrigins causes an error.
func TestRegexEmptyPatternsError(t *testing.T) {
	cfg := cors.Config{
		Mode:         cors.OriginModeRegex,
		RegexOrigins: []string{},
	}
	_, err := cfg.ToMiddleware()
	if err == nil {
		t.Fatal("expected error for empty RegexOrigins, got nil")
	}
}

// TestFuncOriginAllowed verifies that AllowOriginFunc returning allowed=true
// sets ACAO to the returned origin.
func TestFuncOriginAllowed(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode: cors.OriginModeFunc,
		AllowOriginFunc: func(_ *echo.Context, origin string) (string, bool, error) {
			if origin == "https://trusted.com" {
				return origin, true, nil
			}
			return "", false, nil
		},
	})

	c, rec := newContext(e, http.MethodGet, "https://trusted.com")
	if err := mw(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got != "https://trusted.com" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "https://trusted.com")
	}
}

// TestFuncOriginDenied verifies that AllowOriginFunc returning allowed=false
// does NOT set ACAO but still calls next.
func TestFuncOriginDenied(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode: cors.OriginModeFunc,
		AllowOriginFunc: func(_ *echo.Context, origin string) (string, bool, error) {
			return "", false, nil
		},
	})

	nextCalled := false
	next := func(c *echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "ok")
	}

	c, rec := newContext(e, http.MethodGet, "https://denied.com")
	if err := mw(next)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Error("next handler was not called for denied func origin")
	}
	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

// TestFuncOriginError verifies that an error from AllowOriginFunc is propagated.
func TestFuncOriginError(t *testing.T) {
	e := echo.New()
	sentinel := errors.New("origin lookup failed")
	mw := cors.Middleware(cors.Config{
		Mode: cors.OriginModeFunc,
		AllowOriginFunc: func(_ *echo.Context, origin string) (string, bool, error) {
			return "", false, sentinel
		},
	})

	c, _ := newContext(e, http.MethodGet, "https://any.com")
	err := mw(okHandler)(c)
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

// TestFuncNilFuncError verifies that a nil AllowOriginFunc returns an error
// from ToMiddleware.
func TestFuncNilFuncError(t *testing.T) {
	cfg := cors.Config{
		Mode:            cors.OriginModeFunc,
		AllowOriginFunc: nil,
	}
	_, err := cfg.ToMiddleware()
	if err == nil {
		t.Fatal("expected error for nil AllowOriginFunc, got nil")
	}
}

// TestPreflightOPTIONSAllowedOrigin verifies that an OPTIONS request with an
// allowed origin returns 204 No Content.
func TestPreflightOPTIONSAllowedOrigin(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode:         cors.OriginModeExact,
		AllowOrigins: []string{"https://example.com"},
		MaxAge:       3600,
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := mw(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Header().Get("Access-Control-Max-Age") != "3600" {
		t.Errorf("Access-Control-Max-Age = %q, want %q", rec.Header().Get("Access-Control-Max-Age"), "3600")
	}
}

// TestPreflightOPTIONSNoOrigin verifies that an OPTIONS request without an
// Origin header returns 204 without CORS headers (not a CORS request).
func TestPreflightOPTIONSNoOrigin(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode:         cors.OriginModeExact,
		AllowOrigins: []string{"https://example.com"},
	})

	c, rec := newContext(e, http.MethodOptions, "")
	if err := mw(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

// TestAllowCredentialsHeader verifies that AllowCredentials=true sets the
// Access-Control-Allow-Credentials header.
func TestAllowCredentialsHeader(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode:             cors.OriginModeExact,
		AllowOrigins:     []string{"https://example.com"},
		AllowCredentials: true,
	})

	c, rec := newContext(e, http.MethodGet, "https://example.com")
	if err := mw(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := rec.Header().Get("Access-Control-Allow-Credentials")
	if got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
	}
}

// TestWildcardWithCredentialsError verifies that combining "*" with
// AllowCredentials=true causes ToMiddleware to return an error (insecure combo).
func TestWildcardWithCredentialsError(t *testing.T) {
	cfg := cors.Config{
		Mode:             cors.OriginModeExact,
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	}
	_, err := cfg.ToMiddleware()
	if err == nil {
		t.Fatal("expected error for wildcard + credentials, got nil")
	}
}

// TestSkipperBypasses verifies that when Skipper returns true, no CORS headers
// are set and next is called.
func TestSkipperBypasses(t *testing.T) {
	e := echo.New()
	mw := cors.Middleware(cors.Config{
		Mode:         cors.OriginModeExact,
		AllowOrigins: []string{"https://example.com"},
		Skipper:      func(*echo.Context) bool { return true },
	})

	nextCalled := false
	next := func(c *echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "ok")
	}

	c, rec := newContext(e, http.MethodGet, "https://example.com")
	if err := mw(next)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Error("next was not called when skipper returned true")
	}
	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty when skipped", got)
	}
}

// TestUnknownModeError verifies that an unrecognised OriginMode returns an error.
func TestUnknownModeError(t *testing.T) {
	cfg := cors.Config{
		Mode: cors.OriginMode(99),
	}
	_, err := cfg.ToMiddleware()
	if err == nil {
		t.Fatal("expected error for unknown OriginMode, got nil")
	}
}
