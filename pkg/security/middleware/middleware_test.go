package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/security/fetchmeta"
	"github.com/go-sum/security/origin"
	"github.com/labstack/echo/v5"
)

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
