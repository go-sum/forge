package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/security/fetchmeta"
	"github.com/go-sum/security/headers"
	"github.com/go-sum/security/origin"
	"github.com/labstack/echo/v5"
)

func TestSecurityHeadersSetsHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SecurityHeaders(headers.Policy{
		XSSProtection:         "0",
		ContentTypeNosniff:    true,
		FrameOptions:          "DENY",
		ContentSecurityPolicy: "default-src 'self'",
		HSTS: headers.HSTSConfig{
			Enabled:           true,
			MaxAge:            31536000,
			IncludeSubDomains: true,
			Preload:           true,
		},
	})(func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)
	if err != nil {
		t.Fatalf("SecurityHeaders() error = %v", err)
	}

	tests := map[string]string{
		"X-XSS-Protection":          "0",
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Content-Security-Policy":   "default-src 'self'",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
	}
	for name, want := range tests {
		if got := rec.Header().Get(name); got != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestProtectBrowserMutationAllowsVerifiedRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ProtectBrowserMutation(
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
		t.Fatalf("ProtectBrowserMutation() error = %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestProtectBrowserMutationReturnsTypedError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ProtectBrowserMutation(
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
		t.Fatal("ProtectBrowserMutation() error = nil")
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
