package handler

import (
	"net/http"
	"testing"

	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

// TestSafeReverse verifies that route.SafeReverse handles missing names and
// parameterized routes without panicking.
func TestSafeReverse(t *testing.T) {
	e := echo.New()
	noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/users/:id", Name: "user.show", Handler: noOp})
	routes := e.Router().Routes()

	tests := []struct {
		name     string
		wantPath string
		wantOK   bool
	}{
		{"home.show", "/", true},
		{"user.show", "", false},
		{"no.such.route", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := route.SafeReverse(routes, tc.name)
			if ok != tc.wantOK {
				t.Errorf("route.SafeReverse(%q) ok = %v, want %v", tc.name, ok, tc.wantOK)
			}
			if got != tc.wantPath {
				t.Errorf("route.SafeReverse(%q) path = %q, want %q", tc.name, got, tc.wantPath)
			}
		})
	}
}
