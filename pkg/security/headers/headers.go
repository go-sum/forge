// Package headers provides security header utilities that are not covered by
// Echo's built-in Secure middleware. Security headers (XSS protection, CSP,
// HSTS, frame options, etc.) are delegated to middleware.SecureWithConfig;
// this package provides InjectDirectiveSources, which pre-processes a CSP
// string at startup to embed hashes and source allowances into any directive.
package headers

import "strings"

// InjectDirectiveSources prepends sources into the named CSP directive.
// The directive must already be present in csp; if it is not, csp is returned unchanged.
func InjectDirectiveSources(csp, directive string, sources []string) string {
	if csp == "" || len(sources) == 0 {
		return csp
	}

	clean := make([]string, 0, len(sources))
	for _, src := range sources {
		src = strings.TrimSpace(src)
		if src != "" {
			clean = append(clean, src)
		}
	}
	if len(clean) == 0 {
		return csp
	}

	return strings.Replace(csp, directive, directive+" "+strings.Join(clean, " "), 1)
}
