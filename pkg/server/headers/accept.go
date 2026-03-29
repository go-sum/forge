package headers

import (
	"sort"
	"strconv"
	"strings"
)

// LanguageItem is a single parsed entry from an Accept-Language header value.
type LanguageItem struct {
	Tag     string  // e.g. "en-US", "fr", "*"
	Quality float64 // q value in [0.0, 1.0]; default 1.0
}

// AcceptLanguage is a quality-weighted list of language tags parsed from an
// Accept-Language header value, sorted by Quality descending.
type AcceptLanguage []LanguageItem

// ParseAcceptLanguage parses a raw Accept-Language header value and returns a
// quality-sorted AcceptLanguage slice. Items with q=0 are dropped. An
// empty or blank header returns an empty (non-nil) AcceptLanguage.
func ParseAcceptLanguage(header string) AcceptLanguage {
	result := AcceptLanguage{}
	if strings.TrimSpace(header) == "" {
		return result
	}

	for _, token := range strings.Split(header, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		parts := strings.Split(token, ";")
		tag := strings.TrimSpace(parts[0])
		if tag == "" {
			continue
		}

		quality := 1.0
		for _, param := range parts[1:] {
			param = strings.TrimSpace(param)
			lower := strings.ToLower(param)
			if !strings.HasPrefix(lower, "q=") {
				continue
			}
			qStr := param[2:]
			q, err := strconv.ParseFloat(qStr, 64)
			if err != nil {
				// invalid q param — drop the entire item
				quality = -1
				break
			}
			quality = q
			break
		}

		if quality < 0 {
			continue
		}
		if quality == 0.0 {
			continue
		}

		result = append(result, LanguageItem{Tag: tag, Quality: quality})
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Quality > result[j].Quality
	})

	return result
}

// Preferred returns the best matching candidate for the client's language
// preferences. It iterates over the quality-sorted AcceptLanguage and returns
// the first candidate that satisfies any language item via exact
// case-insensitive match or subtag prefix match. Returns "" if no match.
func (a AcceptLanguage) Preferred(candidates []string) string {
	for _, item := range a {
		if item.Tag == "*" {
			for _, c := range candidates {
				if c != "" {
					return c
				}
			}
			continue
		}

		clientLower := strings.ToLower(item.Tag)
		for _, c := range candidates {
			if c == "" {
				continue
			}
			// Exact case-insensitive match.
			if strings.EqualFold(item.Tag, c) {
				return c
			}
			// Subtag prefix match: client "en-US" matches candidate "en".
			if strings.HasPrefix(clientLower, strings.ToLower(c)+"-") {
				return c
			}
		}
	}
	return ""
}

// String serialises the AcceptLanguage back to a header value. Items with
// q==1.0 omit the q parameter. Returns "" for an empty slice.
func (a AcceptLanguage) String() string {
	if len(a) == 0 {
		return ""
	}
	parts := make([]string, 0, len(a))
	for _, item := range a {
		if item.Quality == 1.0 {
			parts = append(parts, item.Tag)
		} else {
			q := strconv.FormatFloat(item.Quality, 'f', -1, 64)
			parts = append(parts, item.Tag+";q="+q)
		}
	}
	return strings.Join(parts, ", ")
}

// ContentItem is a single parsed entry from an Accept header value.
type ContentItem struct {
	Type    string            // e.g. "text/html", "application/*", "*/*"
	Quality float64
	Params  map[string]string // non-q extension parameters
}

// Accept is a quality-weighted list of media types parsed from an Accept
// header value, sorted by Quality descending with more-specific types
// ranked higher at equal quality.
type Accept []ContentItem

// specificity returns 2 for an exact type (no wildcard), 1 for a type/*
// wildcard, and 0 for */*.
func specificity(mediaType string) int {
	if mediaType == "*/*" {
		return 0
	}
	if strings.HasSuffix(mediaType, "/*") {
		return 1
	}
	return 2
}

// ParseAccept parses a raw Accept header value and returns a sorted Accept
// slice. Items with q=0 are dropped. An empty or blank header returns an
// empty (non-nil) Accept.
func ParseAccept(header string) Accept {
	result := Accept{}
	if strings.TrimSpace(header) == "" {
		return result
	}

	for _, token := range strings.Split(header, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		parts := strings.Split(token, ";")
		mediaType := strings.TrimSpace(parts[0])
		if mediaType == "" {
			continue
		}

		quality := 1.0
		params := map[string]string{}
		drop := false

		for _, param := range parts[1:] {
			param = strings.TrimSpace(param)
			if param == "" {
				continue
			}
			lower := strings.ToLower(param)
			if strings.HasPrefix(lower, "q=") {
				qStr := param[2:]
				q, err := strconv.ParseFloat(qStr, 64)
				if err != nil {
					drop = true
					break
				}
				quality = q
				continue
			}
			// Non-q param: split on first "=".
			idx := strings.IndexByte(param, '=')
			if idx < 0 {
				params[strings.TrimSpace(param)] = ""
			} else {
				k := strings.TrimSpace(param[:idx])
				v := strings.TrimSpace(param[idx+1:])
				params[k] = v
			}
		}

		if drop || quality == 0.0 {
			continue
		}

		result = append(result, ContentItem{
			Type:    mediaType,
			Quality: quality,
			Params:  params,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		qi, qj := result[i].Quality, result[j].Quality
		if qi != qj {
			return qi > qj
		}
		return specificity(result[i].Type) > specificity(result[j].Type)
	})

	return result
}

// Preferred returns the best matching candidate for the client's Accept
// preferences. It iterates over the quality-sorted Accept and returns the
// first candidate that satisfies any content item. Returns "" if no match.
func (a Accept) Preferred(candidates []string) string {
	for _, item := range a {
		if item.Type == "*/*" {
			for _, c := range candidates {
				if c != "" {
					return c
				}
			}
			continue
		}

		if strings.HasSuffix(item.Type, "/*") {
			// e.g. "text/*" — match any candidate starting with "text/".
			major := strings.ToLower(item.Type[:len(item.Type)-1]) // "text/"
			for _, c := range candidates {
				if c == "" {
					continue
				}
				if strings.HasPrefix(strings.ToLower(c), major) {
					return c
				}
			}
			continue
		}

		// Exact match.
		for _, c := range candidates {
			if c == "" {
				continue
			}
			if strings.EqualFold(item.Type, c) {
				return c
			}
		}
	}
	return ""
}

// String serialises the Accept back to a header value. Items with q==1.0
// omit the q parameter. Non-q params are appended as ;key=value before q.
// Returns "" for an empty slice.
func (a Accept) String() string {
	if len(a) == 0 {
		return ""
	}
	parts := make([]string, 0, len(a))
	for _, item := range a {
		var sb strings.Builder
		sb.WriteString(item.Type)

		// Append non-q params in a stable order.
		keys := make([]string, 0, len(item.Params))
		for k := range item.Params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteByte(';')
			sb.WriteString(k)
			if v := item.Params[k]; v != "" {
				sb.WriteByte('=')
				sb.WriteString(v)
			}
		}

		if item.Quality < 1.0 {
			sb.WriteString(";q=")
			sb.WriteString(strconv.FormatFloat(item.Quality, 'f', -1, 64))
		}

		parts = append(parts, sb.String())
	}
	return strings.Join(parts, ", ")
}
