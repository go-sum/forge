package headers

import (
	"net/http"
	"strconv"
	"strings"
)

// HSTSConfig configures the Strict-Transport-Security header.
type HSTSConfig struct {
	Enabled           bool
	MaxAge            int
	IncludeSubDomains bool
	Preload           bool
}

// Policy defines reusable security header settings.
type Policy struct {
	XSSProtection         string
	ContentTypeNosniff    bool
	FrameOptions          string
	ContentSecurityPolicy string
	HSTS                  HSTSConfig
}

// Apply writes the configured security headers to h.
func Apply(h http.Header, policy Policy) {
	if policy.XSSProtection != "" {
		h.Set("X-XSS-Protection", policy.XSSProtection)
	}
	if policy.ContentTypeNosniff {
		h.Set("X-Content-Type-Options", "nosniff")
	}
	if policy.FrameOptions != "" {
		h.Set("X-Frame-Options", policy.FrameOptions)
	}
	if policy.ContentSecurityPolicy != "" {
		h.Set("Content-Security-Policy", policy.ContentSecurityPolicy)
	}
	if value := hstsValue(policy.HSTS); value != "" {
		h.Set("Strict-Transport-Security", value)
	}
}

// InjectScriptHashes appends hashes to the script-src directive in csp.
func InjectScriptHashes(csp string, hashes []string) string {
	if csp == "" || len(hashes) == 0 {
		return csp
	}

	clean := make([]string, 0, len(hashes))
	for _, hash := range hashes {
		hash = strings.TrimSpace(hash)
		if hash != "" {
			clean = append(clean, hash)
		}
	}
	if len(clean) == 0 {
		return csp
	}

	return strings.Replace(csp, "script-src", "script-src "+strings.Join(clean, " "), 1)
}

func hstsValue(cfg HSTSConfig) string {
	if !cfg.Enabled || cfg.MaxAge <= 0 {
		return ""
	}

	parts := []string{"max-age=" + strconv.Itoa(cfg.MaxAge)}
	if cfg.IncludeSubDomains {
		parts = append(parts, "includeSubDomains")
	}
	if cfg.Preload {
		parts = append(parts, "preload")
	}
	return strings.Join(parts, "; ")
}
