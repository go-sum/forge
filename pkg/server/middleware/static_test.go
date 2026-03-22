package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

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
