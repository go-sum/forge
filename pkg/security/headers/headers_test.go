package headers

import (
	"net/http"
	"testing"
)

func TestApplySetsConfiguredHeaders(t *testing.T) {
	h := http.Header{}

	Apply(h, Policy{
		XSSProtection:         "0",
		ContentTypeNosniff:    true,
		FrameOptions:          "DENY",
		ContentSecurityPolicy: "default-src 'self'",
		HSTS: HSTSConfig{
			Enabled:           true,
			MaxAge:            31536000,
			IncludeSubDomains: true,
			Preload:           true,
		},
	})

	tests := map[string]string{
		"X-XSS-Protection":          "0",
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Content-Security-Policy":   "default-src 'self'",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
	}
	for name, want := range tests {
		if got := h.Get(name); got != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestInjectScriptHashes(t *testing.T) {
	got := InjectScriptHashes("default-src 'self'; script-src 'self'; style-src 'self'", []string{"'sha256-a'", "", " 'sha256-b' "})
	want := "default-src 'self'; script-src 'sha256-a' 'sha256-b' 'self'; style-src 'self'"
	if got != want {
		t.Fatalf("InjectScriptHashes() = %q, want %q", got, want)
	}
}

func TestInjectScriptHashesLeavesUntouchedWhenNoHashes(t *testing.T) {
	csp := "default-src 'self'; script-src 'self'"
	if got := InjectScriptHashes(csp, nil); got != csp {
		t.Fatalf("InjectScriptHashes() = %q, want %q", got, csp)
	}
}
