package cache_test

import (
	"fmt"
	"hash/crc32"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-sum/server/cache"
)

func TestWeakETag(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    string
	}{
		{
			name:    "empty content",
			content: []byte{},
			want:    fmt.Sprintf(`W/"0-%x"`, crc32.ChecksumIEEE([]byte{})),
		},
		{
			name:    "known content",
			content: []byte("hello"),
			want:    fmt.Sprintf(`W/"5-%x"`, crc32.ChecksumIEEE([]byte("hello"))),
		},
		{
			name:    "deterministic — same input same output",
			content: []byte("deterministic"),
			want:    fmt.Sprintf(`W/"13-%x"`, crc32.ChecksumIEEE([]byte("deterministic"))),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cache.WeakETag(tc.content)
			if got != tc.want {
				t.Errorf("WeakETag() = %q, want %q", got, tc.want)
			}
			// Determinism: call twice, expect same result.
			if cache.WeakETag(tc.content) != got {
				t.Error("WeakETag() is not deterministic")
			}
		})
	}
}

func TestStrongETag(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		// SHA-256 of empty string: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
		wantPrefix string
		wantLen    int
	}{
		{
			name:       "empty content",
			content:    []byte{},
			wantPrefix: `"e3b0c44298fc1c149afbf4c8996fb924`,
			wantLen:    66, // 2 quotes + 64 hex chars
		},
		{
			name:       "non-empty content produces 66-char ETag",
			content:    []byte("hello"),
			wantLen:    66,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cache.StrongETag(tc.content)
			if len(got) != tc.wantLen {
				t.Errorf("StrongETag() length = %d, want %d; got %q", len(got), tc.wantLen, got)
			}
			if got[0] != '"' || got[len(got)-1] != '"' {
				t.Errorf("StrongETag() not wrapped in quotes: %q", got)
			}
			if tc.wantPrefix != "" && got[:len(tc.wantPrefix)] != tc.wantPrefix {
				t.Errorf("StrongETag() prefix = %q, want %q", got[:len(tc.wantPrefix)], tc.wantPrefix)
			}
			// Determinism
			if cache.StrongETag(tc.content) != got {
				t.Error("StrongETag() is not deterministic")
			}
		})
	}
}

func TestSetETag(t *testing.T) {
	h := make(http.Header)
	cache.SetETag(h, `W/"42-abc"`)
	if got := h.Get("ETag"); got != `W/"42-abc"` {
		t.Errorf("SetETag() header = %q, want %q", got, `W/"42-abc"`)
	}
}

func TestSetLastModified(t *testing.T) {
	h := make(http.Header)
	ts := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	cache.SetLastModified(h, ts)
	want := ts.UTC().Format(http.TimeFormat)
	if got := h.Get("Last-Modified"); got != want {
		t.Errorf("SetLastModified() header = %q, want %q", got, want)
	}
}

func TestCheckIfNoneMatch(t *testing.T) {
	tests := []struct {
		name   string
		header string
		etag   string
		want   bool
	}{
		{name: "absent header", header: "", etag: `W/"5-abc"`, want: false},
		{name: "wildcard", header: "*", etag: `W/"5-abc"`, want: true},
		{name: "exact single match", header: `W/"5-abc"`, etag: `W/"5-abc"`, want: true},
		{name: "exact single no match", header: `W/"5-xyz"`, etag: `W/"5-abc"`, want: false},
		{name: "comma list contains match", header: `W/"1-a", W/"5-abc", W/"9-z"`, etag: `W/"5-abc"`, want: true},
		{name: "comma list no match", header: `W/"1-a", W/"2-b"`, etag: `W/"5-abc"`, want: false},
		{name: "comma list with spaces", header: `  W/"5-abc"  , W/"9-z"`, etag: `W/"5-abc"`, want: true},
		{name: "case sensitive — no match on different case", header: `W/"5-ABC"`, etag: `W/"5-abc"`, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("If-None-Match", tc.header)
			}
			got := cache.CheckIfNoneMatch(req, tc.etag)
			if got != tc.want {
				t.Errorf("CheckIfNoneMatch() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCheckIfModifiedSince(t *testing.T) {
	base := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		header string      // raw If-Modified-Since value ("" = absent)
		t      time.Time   // resource modification time
		want   bool
	}{
		{name: "absent header", header: "", t: base, want: false},
		{name: "malformed header", header: "not-a-date", t: base, want: false},
		{
			name:   "IMS equals t — not modified",
			header: base.UTC().Format(http.TimeFormat),
			t:      base,
			want:   true,
		},
		{
			name:   "IMS after t — not modified (IMS is newer than resource)",
			header: base.Add(time.Hour).UTC().Format(http.TimeFormat),
			t:      base,
			want:   true,
		},
		{
			name:   "t after IMS — modified (resource is newer than IMS)",
			header: base.Add(-time.Hour).UTC().Format(http.TimeFormat),
			t:      base,
			want:   false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("If-Modified-Since", tc.header)
			}
			got := cache.CheckIfModifiedSince(req, tc.t)
			if got != tc.want {
				t.Errorf("CheckIfModifiedSince() = %v, want %v", got, tc.want)
			}
		})
	}
}
