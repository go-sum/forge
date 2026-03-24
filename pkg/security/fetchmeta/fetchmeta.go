package fetchmeta

import (
	"net/http"
	"slices"
)

// Policy defines Fetch Metadata validation settings for unsafe requests.
type Policy struct {
	Enabled                 bool
	AllowedSites            []string
	AllowedModes            []string
	AllowedDestinations     []string
	FallbackWhenMissing     bool
	RejectCrossSiteNavigate bool
}

// Result describes the validation outcome.
type Result struct {
	Valid          bool
	Reason         string
	HeadersMissing bool
}

// Validate checks Fetch Metadata request headers.
func Validate(r *http.Request, policy Policy) Result {
	if !policy.Enabled {
		return Result{Valid: true}
	}

	site := r.Header.Get("Sec-Fetch-Site")
	mode := r.Header.Get("Sec-Fetch-Mode")
	dest := r.Header.Get("Sec-Fetch-Dest")

	if site == "" && mode == "" && dest == "" {
		if policy.FallbackWhenMissing {
			return Result{Valid: true, HeadersMissing: true}
		}
		return Result{
			Reason:         "fetch metadata headers are required",
			HeadersMissing: true,
		}
	}

	if len(policy.AllowedSites) > 0 && site != "" && !slices.Contains(policy.AllowedSites, site) {
		return Result{Reason: "sec-fetch-site is not allowed"}
	}
	if len(policy.AllowedModes) > 0 && mode != "" && !slices.Contains(policy.AllowedModes, mode) {
		return Result{Reason: "sec-fetch-mode is not allowed"}
	}
	if len(policy.AllowedDestinations) > 0 && dest != "" && !slices.Contains(policy.AllowedDestinations, dest) {
		return Result{Reason: "sec-fetch-dest is not allowed"}
	}
	if policy.RejectCrossSiteNavigate && site == "cross-site" && mode == "navigate" {
		return Result{Reason: "cross-site navigate request is not allowed"}
	}

	return Result{Valid: true}
}
