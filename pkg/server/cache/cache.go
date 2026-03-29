// Package cache provides ETag generation and HTTP conditional-request helpers
// for use with response caching middleware.
//
// All functions depend only on the standard library and are safe to use from
// any pkg/ package or application layer.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"net/http"
	"strings"
	"time"
)

// WeakETag returns a weak ETag of the form W/"<len>-<crc32hex>".
// It uses the content length and IEEE CRC32 checksum — fast and non-cryptographic.
// Suitable for fragment caching where exact byte-identity is not required.
func WeakETag(content []byte) string {
	h := crc32.ChecksumIEEE(content)
	return fmt.Sprintf(`W/"%d-%x"`, len(content), h)
}

// StrongETag returns a strong ETag of the form "<sha256hex>".
// It uses a full SHA-256 hash — cryptographically strong and suitable when
// exact byte-identity must be guaranteed.
func StrongETag(content []byte) string {
	sum := sha256.Sum256(content)
	return `"` + hex.EncodeToString(sum[:]) + `"`
}

// SetETag writes the ETag response header. tag must already be a correctly
// formatted ETag value (e.g. the return value of WeakETag or StrongETag).
func SetETag(h http.Header, tag string) {
	h.Set("ETag", tag)
}

// SetLastModified writes the Last-Modified response header using the HTTP
// date format defined by RFC 7231.
func SetLastModified(h http.Header, t time.Time) {
	h.Set("Last-Modified", t.UTC().Format(http.TimeFormat))
}

// CheckIfNoneMatch reports whether the request's If-None-Match header matches
// etag, indicating that a 304 Not Modified response is appropriate.
//
// It handles the wildcard value "*" and a comma-separated list of quoted ETag
// strings. Comparison is exact (ETags are opaque strings).
func CheckIfNoneMatch(r *http.Request, etag string) bool {
	inm := strings.TrimSpace(r.Header.Get("If-None-Match"))
	if inm == "" {
		return false
	}
	if inm == "*" {
		return true
	}
	for _, candidate := range strings.Split(inm, ",") {
		if strings.TrimSpace(candidate) == etag {
			return true
		}
	}
	return false
}

// CheckIfModifiedSince reports whether the content has NOT been modified since
// the time in the request's If-Modified-Since header, indicating that a 304
// Not Modified response is appropriate.
//
// Returns false (treat as modified) if the header is absent or unparseable.
func CheckIfModifiedSince(r *http.Request, t time.Time) bool {
	ims := r.Header.Get("If-Modified-Since")
	if ims == "" {
		return false
	}
	parsed, err := http.ParseTime(ims)
	if err != nil {
		return false
	}
	return !t.After(parsed)
}
