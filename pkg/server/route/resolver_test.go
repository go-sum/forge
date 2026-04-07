package route_test

import (
	"net/http"
	"testing"

	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

func newTestEcho() (*echo.Echo, func(c *echo.Context) error) {
	e := echo.New()
	noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }
	return e, noOp
}

func TestResolver_Path(t *testing.T) {
	e, noOp := newTestEcho()
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: noOp})

	resolve := route.NewResolver(func() echo.Routes { return e.Router().Routes() })

	tests := []struct {
		routeName string
		want      string
	}{
		{"home.show", "/"},
		{"signin.get", "/signin"},
	}

	for _, tc := range tests {
		t.Run(tc.routeName, func(t *testing.T) {
			fn := resolve.Path(tc.routeName)
			got := fn()
			if got != tc.want {
				t.Errorf("Path(%q)() = %q, want %q", tc.routeName, got, tc.want)
			}
		})
	}
}

func TestResolver_Path_WithParams(t *testing.T) {
	e, noOp := newTestEcho()
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/users/:id", Name: "user.show", Handler: noOp})

	resolve := route.NewResolver(func() echo.Routes { return e.Router().Routes() })

	fn := resolve.Path("user.show", "abc-123")
	got := fn()
	want := "/users/abc-123"
	if got != want {
		t.Errorf("Path(%q, %q)() = %q, want %q", "user.show", "abc-123", got, want)
	}
}

func TestResolver_Path_UnknownPanics(t *testing.T) {
	e, _ := newTestEcho()
	resolve := route.NewResolver(func() echo.Routes { return e.Router().Routes() })

	fn := resolve.Path("no.such.route")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unknown route name, got none")
		}
	}()
	fn()
}

func TestResolver_URL(t *testing.T) {
	e, noOp := newTestEcho()
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: noOp})

	resolve := route.NewResolver(func() echo.Routes { return e.Router().Routes() })

	fn := resolve.URL("https://example.com", "signin.get")
	got := fn()
	want := "https://example.com/signin"
	if got != want {
		t.Errorf("URL(%q, %q)() = %q, want %q", "https://example.com", "signin.get", got, want)
	}
}

func TestResolver_Routes(t *testing.T) {
	e, noOp := newTestEcho()
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: noOp})

	resolve := route.NewResolver(func() echo.Routes { return e.Router().Routes() })

	routes := resolve.Routes()()
	if len(routes) == 0 {
		t.Error("Routes()() returned empty route list, want at least one route")
	}
}

func TestResolver_Lazy(t *testing.T) {
	e, noOp := newTestEcho()

	// Create resolver before registering routes.
	resolve := route.NewResolver(func() echo.Routes { return e.Router().Routes() })

	// Register routes after resolver is created.
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/late", Name: "late.get", Handler: noOp})

	fn := resolve.Path("late.get")
	got := fn()
	want := "/late"
	if got != want {
		t.Errorf("lazy Path(%q)() = %q, want %q", "late.get", got, want)
	}
}
