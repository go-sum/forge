package fetchmeta

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateAllowsSameOriginRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/users", nil)
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")

	got := Validate(req, Policy{
		Enabled:                 true,
		AllowedSites:            []string{"same-origin", "same-site"},
		AllowedModes:            []string{"cors", "navigate", "same-origin"},
		FallbackWhenMissing:     true,
		RejectCrossSiteNavigate: true,
	})
	if !got.Valid {
		t.Fatalf("Validate() = %#v", got)
	}
}

func TestValidateAllowsMissingHeadersWhenFallbackEnabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/users", nil)

	got := Validate(req, Policy{
		Enabled:             true,
		FallbackWhenMissing: true,
	})
	if !got.Valid || !got.HeadersMissing {
		t.Fatalf("Validate() = %#v", got)
	}
}

func TestValidateRejectsMissingHeadersWhenFallbackDisabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/users", nil)

	got := Validate(req, Policy{
		Enabled:             true,
		FallbackWhenMissing: false,
	})
	if got.Valid || !got.HeadersMissing {
		t.Fatalf("Validate() = %#v", got)
	}
}

func TestValidateRejectsCrossSiteNavigate(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/users", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Sec-Fetch-Mode", "navigate")

	got := Validate(req, Policy{
		Enabled:                 true,
		AllowedSites:            []string{"same-origin", "same-site", "cross-site"},
		AllowedModes:            []string{"cors", "navigate"},
		FallbackWhenMissing:     true,
		RejectCrossSiteNavigate: true,
	})
	if got.Valid {
		t.Fatalf("Validate() = %#v, want invalid", got)
	}
}

func TestValidateRejectsDisallowedSite(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://internal/users", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	got := Validate(req, Policy{
		Enabled:             true,
		AllowedSites:        []string{"same-origin", "same-site"},
		FallbackWhenMissing: true,
	})
	if got.Valid {
		t.Fatalf("Validate() = %#v, want invalid", got)
	}
}
