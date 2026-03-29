package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestCacheHeadersSetsCacheControlAndVary(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := CacheHeaders("private, must-revalidate", "Cookie")(func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)
	if err != nil {
		t.Fatalf("CacheHeaders() error = %v", err)
	}
	if got := rec.Header().Get("Cache-Control"); got != "private, must-revalidate" {
		t.Fatalf("Cache-Control = %q, want %q", got, "private, must-revalidate")
	}
	if got := rec.Header().Get("Vary"); got != "Cookie" {
		t.Fatalf("Vary = %q, want %q", got, "Cookie")
	}
}

func TestCacheHeadersAppendsVaryAndSkipsEmptyCacheControl(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().Header().Set("Vary", "Accept-Encoding")

	err := CacheHeaders("", "Cookie")(func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)
	if err != nil {
		t.Fatalf("CacheHeaders() error = %v", err)
	}
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("Cache-Control = %q, want empty", got)
	}
	if got := rec.Header().Get("Vary"); got != "Accept-Encoding, Cookie" {
		t.Fatalf("Vary = %q, want %q", got, "Accept-Encoding, Cookie")
	}
}
