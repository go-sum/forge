package headers

import (
	"sort"
	"strconv"
	"strings"
)

// CacheControl holds the directives parsed from a Cache-Control header value.
// Directive names are stored lower-cased. Flag directives (no value) are stored
// with an empty string value. Unknown directives are retained and accessible
// via [CacheControl.Has].
type CacheControl struct {
	directives map[string]string // lower-cased name → value ("" for flags)
}

// ParseCacheControl parses a raw Cache-Control header value and returns a
// CacheControl. Directive names are lower-cased; the first occurrence of a
// duplicate directive wins. An empty or blank header returns a zero CacheControl.
func ParseCacheControl(header string) CacheControl {
	if strings.TrimSpace(header) == "" {
		return CacheControl{}
	}

	directives := map[string]string{}

	for _, token := range strings.Split(header, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		// Split on first "=" only.
		idx := strings.IndexByte(token, '=')
		var name, value string
		if idx < 0 {
			name = strings.ToLower(strings.TrimSpace(token))
			value = ""
		} else {
			name = strings.ToLower(strings.TrimSpace(token[:idx]))
			value = strings.TrimSpace(token[idx+1:])
			// Strip surrounding quotes if present.
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
		}

		if name == "" {
			continue
		}

		// First occurrence wins.
		if _, exists := directives[name]; !exists {
			directives[name] = value
		}
	}

	return CacheControl{directives: directives}
}

// maxAgeDirective is a shared helper for MaxAge and SMaxAge.
func (c CacheControl) maxAgeDirective(name string) (seconds int, ok bool) {
	v, exists := c.directives[name]
	if !exists {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
}

// MaxAge returns the max-age directive value in seconds and whether it was present.
// Returns (0, false) if absent or if the value is not a valid integer.
func (c CacheControl) MaxAge() (seconds int, ok bool) {
	return c.maxAgeDirective("max-age")
}

// SMaxAge returns the s-maxage directive value in seconds and whether it was present.
// Returns (0, false) if absent or if the value is not a valid integer.
func (c CacheControl) SMaxAge() (seconds int, ok bool) {
	return c.maxAgeDirective("s-maxage")
}

// NoStore reports whether the no-store directive is set.
func (c CacheControl) NoStore() bool {
	_, ok := c.directives["no-store"]
	return ok
}

// NoCache reports whether the no-cache directive is set.
func (c CacheControl) NoCache() bool {
	_, ok := c.directives["no-cache"]
	return ok
}

// Private reports whether the private directive is set.
func (c CacheControl) Private() bool {
	_, ok := c.directives["private"]
	return ok
}

// Public reports whether the public directive is set.
func (c CacheControl) Public() bool {
	_, ok := c.directives["public"]
	return ok
}

// MustRevalidate reports whether the must-revalidate directive is set.
func (c CacheControl) MustRevalidate() bool {
	_, ok := c.directives["must-revalidate"]
	return ok
}

// Immutable reports whether the immutable directive is set.
func (c CacheControl) Immutable() bool {
	_, ok := c.directives["immutable"]
	return ok
}

// Has reports whether the named directive (case-insensitive) is present,
// including unknown directives not covered by the typed accessors.
func (c CacheControl) Has(directive string) bool {
	_, ok := c.directives[strings.ToLower(directive)]
	return ok
}

// String serialises the CacheControl directives back to a Cache-Control
// header value. Directives are output in sorted order for deterministic output.
// Flag directives (no value) appear as the directive name alone.
// Valued directives appear as "name=value".
func (c CacheControl) String() string {
	if len(c.directives) == 0 {
		return ""
	}

	keys := make([]string, 0, len(c.directives))
	for k := range c.directives {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := c.directives[k]
		if v == "" {
			parts = append(parts, k)
		} else {
			parts = append(parts, k+"="+v)
		}
	}

	return strings.Join(parts, ", ")
}

// Builder builds a Cache-Control header value programmatically.
// Directives are output in the order they were added.
// Calling the same directive setter more than once is a no-op after the first call.
type Builder struct {
	parts []string
	seen  map[string]struct{}
}

// NewCacheControl returns a new Builder.
func NewCacheControl() *Builder {
	return &Builder{
		seen: map[string]struct{}{},
	}
}

// add is an internal helper that appends a directive to the builder if not
// already present.
func (b *Builder) add(key, value string) *Builder {
	if _, exists := b.seen[key]; exists {
		return b
	}
	b.seen[key] = struct{}{}
	if value == "" {
		b.parts = append(b.parts, key)
	} else {
		b.parts = append(b.parts, key+"="+value)
	}
	return b
}

// MaxAge adds the "max-age=N" directive.
func (b *Builder) MaxAge(seconds int) *Builder {
	return b.add("max-age", strconv.Itoa(seconds))
}

// SMaxAge adds the "s-maxage=N" directive.
func (b *Builder) SMaxAge(seconds int) *Builder {
	return b.add("s-maxage", strconv.Itoa(seconds))
}

// NoStore adds the "no-store" directive.
func (b *Builder) NoStore() *Builder {
	return b.add("no-store", "")
}

// NoCache adds the "no-cache" directive.
func (b *Builder) NoCache() *Builder {
	return b.add("no-cache", "")
}

// Private adds the "private" directive.
func (b *Builder) Private() *Builder {
	return b.add("private", "")
}

// Public adds the "public" directive.
func (b *Builder) Public() *Builder {
	return b.add("public", "")
}

// MustRevalidate adds the "must-revalidate" directive.
func (b *Builder) MustRevalidate() *Builder {
	return b.add("must-revalidate", "")
}

// Immutable adds the "immutable" directive.
func (b *Builder) Immutable() *Builder {
	return b.add("immutable", "")
}

// String returns the final Cache-Control header value.
// Directives appear in the order they were added, joined by ", ".
func (b *Builder) String() string {
	return strings.Join(b.parts, ", ")
}
