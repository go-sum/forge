// Package headers provides security header utilities that are not covered by
// Echo's built-in Secure middleware. Security headers (XSS protection, CSP,
// HSTS, frame options, etc.) are delegated to middleware.SecureWithConfig;
// this package retains only InjectScriptHashes, which pre-processes a CSP
// string at startup to embed inline-script hashes.
package headers

import "strings"

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
