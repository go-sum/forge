package origin

import (
	"net/http"
	"net/url"
	"strings"
)

// Policy defines origin/referer validation settings for unsafe requests.
type Policy struct {
	Enabled         bool
	CanonicalOrigin string
	RequireHeader   bool
	AllowedOrigins  []string
}

// Result describes the validation outcome.
type Result struct {
	Valid          bool
	Reason         string
	Source         string
	HeadersMissing bool
}

// Validate checks Origin and Referer against the canonical origin.
func Validate(r *http.Request, policy Policy) Result {
	if !policy.Enabled {
		return Result{Valid: true}
	}

	expected := normalize(policy.CanonicalOrigin)
	if expected == "" {
		return Result{Reason: "canonical origin is invalid"}
	}

	if origin := r.Header.Get("Origin"); origin != "" {
		if !matches(origin, expected, policy.AllowedOrigins) {
			return Result{
				Reason: "origin header does not match the canonical origin",
				Source: "Origin",
			}
		}
		return Result{Valid: true, Source: "Origin"}
	}

	if referer := r.Header.Get("Referer"); referer != "" {
		refOrigin := originFromURL(referer)
		if refOrigin == "" {
			return Result{
				Reason: "referer header is invalid",
				Source: "Referer",
			}
		}
		if !matches(refOrigin, expected, policy.AllowedOrigins) {
			return Result{
				Reason: "referer header does not match the canonical origin",
				Source: "Referer",
			}
		}
		return Result{Valid: true, Source: "Referer"}
	}

	if policy.RequireHeader {
		return Result{
			Reason:         "origin or referer header is required",
			HeadersMissing: true,
		}
	}

	return Result{Valid: true, HeadersMissing: true}
}

func matches(value string, expected string, allowed []string) bool {
	normalized := normalize(value)
	if normalized == "" {
		return false
	}
	if normalized == expected {
		return true
	}
	for _, item := range allowed {
		if normalize(item) == normalized {
			return true
		}
	}
	return false
}

func normalize(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}

	host := strings.ToLower(u.Hostname())
	port := u.Port()
	switch {
	case port == "":
	case u.Scheme == "http" && port == "80":
		port = ""
	case u.Scheme == "https" && port == "443":
		port = ""
	}
	if port != "" {
		host += ":" + port
	}

	return strings.ToLower(u.Scheme) + "://" + host
}

func originFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}
