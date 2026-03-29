package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestStaticCacheDefaultConfig(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantHeader string
	}{
		{
			name:       "versioned asset gets immutable header",
			target:     "/app.css?v=abc12345",
			wantHeader: cacheImmutable,
		},
		{
			name:       "unversioned asset gets no-cache header",
			target:     "/app.css",
			wantHeader: "no-cache",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mw := StaticCache(StaticCacheConfig{})
			err := mw(func(c *echo.Context) error {
				return c.NoContent(http.StatusOK)
			})(c)
			if err != nil {
				t.Fatalf("StaticCache() error = %v", err)
			}
			if got := rec.Header().Get("Cache-Control"); got != tc.wantHeader {
				t.Fatalf("Cache-Control = %q, want %q", got, tc.wantHeader)
			}
		})
	}
}

func TestStaticCacheCustomVersionParam(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/app.css?hash=abc12345", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := StaticCache(StaticCacheConfig{VersionParam: "hash"})
	err := mw(func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)
	if err != nil {
		t.Fatalf("StaticCache() error = %v", err)
	}
	if got := rec.Header().Get("Cache-Control"); got != cacheImmutable {
		t.Fatalf("Cache-Control = %q, want %q", got, cacheImmutable)
	}
}

func TestStaticCacheCustomHeaders(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantHeader string
	}{
		{
			name:       "versioned asset gets custom versioned header",
			target:     "/app.css?v=abc12345",
			wantHeader: "public, max-age=3600",
		},
		{
			name:       "unversioned asset gets custom unversioned header",
			target:     "/app.css",
			wantHeader: "private, no-store",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mw := StaticCache(StaticCacheConfig{
				VersionedHeader:   "public, max-age=3600",
				UnversionedHeader: "private, no-store",
			})
			err := mw(func(c *echo.Context) error {
				return c.NoContent(http.StatusOK)
			})(c)
			if err != nil {
				t.Fatalf("StaticCache() error = %v", err)
			}
			if got := rec.Header().Get("Cache-Control"); got != tc.wantHeader {
				t.Fatalf("Cache-Control = %q, want %q", got, tc.wantHeader)
			}
		})
	}
}

func TestStaticCacheSkipper(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/app.css?v=abc12345", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := StaticCache(StaticCacheConfig{
		Skipper: func(*echo.Context) bool { return true },
	})
	err := mw(func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)
	if err != nil {
		t.Fatalf("StaticCache() error = %v", err)
	}
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("Cache-Control = %q, want empty (skipped)", got)
	}
}

func TestStaticCacheControlSetsExpectedHeaders(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		prefix     string
		wantHeader string
	}{
		{
			name:       "versioned asset is immutable",
			target:     "/public/app.css?v=abc12345",
			prefix:     "/public",
			wantHeader: cacheImmutable,
		},
		{
			name:       "unversioned asset is no-cache",
			target:     "/public/app.css",
			prefix:     "/public",
			wantHeader: "no-cache",
		},
		{
			name:       "non-matching path is untouched",
			target:     "/assets/app.css",
			prefix:     "/public",
			wantHeader: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := StaticCacheControl(tc.prefix)(func(c *echo.Context) error {
				return c.NoContent(http.StatusOK)
			})(c)
			if err != nil {
				t.Fatalf("StaticCacheControl() error = %v", err)
			}
			if got := rec.Header().Get("Cache-Control"); got != tc.wantHeader {
				t.Fatalf("Cache-Control = %q", got)
			}
		})
	}
}
