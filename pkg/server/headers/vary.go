package headers

import (
	"net/http"
	"strings"
)

// AppendVary adds each value in values to the Vary header of h without
// creating duplicates. Comparison is case-insensitive. Uses [http.Header.Set]
// (not Add) to avoid the duplicate entries that arise when multiple middleware
// each call h.Add("Vary", ...) independently.
//
// If values is empty or all values are already present, h is unchanged.
func AppendVary(h http.Header, values ...string) {
	existing := h.Get("Vary")

	// Build a dedup set from the existing Vary value, and collect existing
	// entries preserving their original casing.
	seen := map[string]struct{}{}
	var merged []string

	if existing != "" {
		for entry := range strings.SplitSeq(existing, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			lower := strings.ToLower(entry)
			if _, dup := seen[lower]; !dup {
				seen[lower] = struct{}{}
				merged = append(merged, entry)
			}
		}
	}

	added := false
	for _, value := range values {
		lower := strings.ToLower(value)
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		merged = append(merged, value)
		added = true
	}

	if !added {
		return
	}

	h.Set("Vary", strings.Join(merged, ", "))
}
