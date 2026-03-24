package origin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateAcceptsCanonicalOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/signin", nil)
	req.Header.Set("Origin", "https://example.com")

	got := Validate(req, Policy{
		Enabled:         true,
		CanonicalOrigin: "https://example.com",
		RequireHeader:   true,
	})
	if !got.Valid {
		t.Fatalf("Validate() = %#v", got)
	}
}

func TestValidateAcceptsAllowedOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/signin", nil)
	req.Header.Set("Origin", "https://admin.example.com")

	got := Validate(req, Policy{
		Enabled:         true,
		CanonicalOrigin: "https://example.com",
		RequireHeader:   true,
		AllowedOrigins:  []string{"https://admin.example.com"},
	})
	if !got.Valid {
		t.Fatalf("Validate() = %#v", got)
	}
}

func TestValidateFallsBackToReferer(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/signin", nil)
	req.Header.Set("Referer", "https://example.com/signin")

	got := Validate(req, Policy{
		Enabled:         true,
		CanonicalOrigin: "https://example.com",
		RequireHeader:   true,
	})
	if !got.Valid || got.Source != "Referer" {
		t.Fatalf("Validate() = %#v", got)
	}
}

func TestValidateRejectsMismatchedOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/signin", nil)
	req.Header.Set("Origin", "https://evil.example")

	got := Validate(req, Policy{
		Enabled:         true,
		CanonicalOrigin: "https://example.com",
		RequireHeader:   true,
	})
	if got.Valid {
		t.Fatalf("Validate() = %#v, want invalid", got)
	}
}

func TestValidateRejectsMissingHeadersWhenRequired(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/signin", nil)

	got := Validate(req, Policy{
		Enabled:         true,
		CanonicalOrigin: "https://example.com",
		RequireHeader:   true,
	})
	if got.Valid || !got.HeadersMissing {
		t.Fatalf("Validate() = %#v, want missing-header failure", got)
	}
}

func TestValidateNormalizesDefaultHTTPSPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/signin", nil)
	req.Header.Set("Origin", "https://example.com:443")

	got := Validate(req, Policy{
		Enabled:         true,
		CanonicalOrigin: "https://example.com",
		RequireHeader:   true,
	})
	if !got.Valid {
		t.Fatalf("Validate() = %#v", got)
	}
}
