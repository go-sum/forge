package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/server/route"
	"github.com/labstack/echo/v5"
)

// headerMW returns middleware that sets key=value on the response, used to
// assert which middleware chain a route is part of.
func headerMW(key, value string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Response().Header().Set(key, value)
			return next(c)
		}
	}
}

func ok(c *echo.Context) error { return c.NoContent(http.StatusOK) }

func do(t *testing.T, e *echo.Echo, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestRegister_Empty(t *testing.T) {
	e := echo.New()
	route.Register(e) // no nodes — must not panic
	rec := do(t, e, http.MethodGet, "/")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestRegister_SingleRoute(t *testing.T) {
	e := echo.New()
	route.Register(e,
		route.Route(http.MethodGet, "/hello", "hello", ok),
	)
	rec := do(t, e, http.MethodGet, "/hello")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestRegister_MethodShortcuts(t *testing.T) {
	e := echo.New()
	route.Register(e,
		route.GET("/get", "r.get", ok),
		route.POST("/post", "r.post", ok),
		route.PUT("/put", "r.put", ok),
		route.DELETE("/delete", "r.delete", ok),
	)

	tests := []struct{ method, path string }{
		{http.MethodGet, "/get"},
		{http.MethodPost, "/post"},
		{http.MethodPut, "/put"},
		{http.MethodDelete, "/delete"},
	}
	for _, tc := range tests {
		rec := do(t, e, tc.method, tc.path)
		if rec.Code != http.StatusOK {
			t.Errorf("%s %s = %d, want 200", tc.method, tc.path, rec.Code)
		}
	}
}

func TestRegister_Group_Prefix(t *testing.T) {
	e := echo.New()
	route.Register(e,
		route.Group("/api",
			route.GET("/users", "user.list", ok),
		),
	)
	if rec := do(t, e, http.MethodGet, "/api/users"); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	// Without prefix should 404.
	if rec := do(t, e, http.MethodGet, "/users"); rec.Code != http.StatusNotFound {
		t.Fatalf("/users status = %d, want 404", rec.Code)
	}
}

func TestRegister_NestedGroups(t *testing.T) {
	e := echo.New()
	route.Register(e,
		route.Group("/api",
			route.Group("/v1",
				route.Group("/users",
					route.GET("", "user.list", ok),
				),
			),
		),
	)
	rec := do(t, e, http.MethodGet, "/api/v1/users")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestRegister_Layout_NoPrefix(t *testing.T) {
	e := echo.New()
	route.Register(e,
		route.Layout(
			route.GET("/a", "a", ok),
			route.GET("/b", "b", ok),
		),
	)
	// Both routes registered at their own paths, not nested under any prefix.
	for _, path := range []string{"/a", "/b"} {
		if rec := do(t, e, http.MethodGet, path); rec.Code != http.StatusOK {
			t.Errorf("%s status = %d, want 200", path, rec.Code)
		}
	}
}

func TestRegister_LayoutWithoutUse_NoCostGroup(t *testing.T) {
	// A Layout with no Use() should not add any middleware to routes.
	e := echo.New()
	called := false
	mw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error { called = true; return next(c) }
	}
	_ = mw // ensure it's defined but NOT used here
	route.Register(e,
		route.Layout( // no Use() — structural grouping only
			route.GET("/ping", "ping", ok),
		),
	)
	rec := do(t, e, http.MethodGet, "/ping")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if called {
		t.Fatal("unexpected middleware call on layout without Use()")
	}
}

func TestRegister_Use_Middleware(t *testing.T) {
	e := echo.New()
	route.Register(e,
		route.Layout(
			route.Use(headerMW("X-Auth", "required")),
			route.GET("/protected", "protected", ok),
		),
		route.GET("/public", "public", ok),
	)

	protectedRec := do(t, e, http.MethodGet, "/protected")
	if protectedRec.Header().Get("X-Auth") != "required" {
		t.Errorf("/protected: X-Auth header missing")
	}
	publicRec := do(t, e, http.MethodGet, "/public")
	if publicRec.Header().Get("X-Auth") != "" {
		t.Errorf("/public: unexpected X-Auth header")
	}
}

func TestRegister_NestedLayout_MiddlewareInheritance(t *testing.T) {
	e := echo.New()
	route.Register(e,
		route.Layout(
			route.Use(headerMW("X-Outer", "yes")),
			route.Layout(
				route.Use(headerMW("X-Inner", "yes")),
				route.GET("/deep", "deep", ok),
			),
			route.GET("/shallow", "shallow", ok),
		),
	)

	deepRec := do(t, e, http.MethodGet, "/deep")
	if deepRec.Header().Get("X-Outer") != "yes" {
		t.Error("/deep: missing X-Outer")
	}
	if deepRec.Header().Get("X-Inner") != "yes" {
		t.Error("/deep: missing X-Inner")
	}

	shallowRec := do(t, e, http.MethodGet, "/shallow")
	if shallowRec.Header().Get("X-Outer") != "yes" {
		t.Error("/shallow: missing X-Outer")
	}
	if shallowRec.Header().Get("X-Inner") != "" {
		t.Error("/shallow: unexpected X-Inner")
	}
}

func TestRegister_Group_WithMiddleware(t *testing.T) {
	e := echo.New()
	route.Register(e,
		route.Group("/admin",
			route.Use(headerMW("X-Admin", "1")),
			route.GET("/users", "admin.users", ok),
			route.GET("/settings", "admin.settings", ok),
		),
		route.GET("/public", "public", ok),
	)

	for _, path := range []string{"/admin/users", "/admin/settings"} {
		rec := do(t, e, http.MethodGet, path)
		if rec.Header().Get("X-Admin") != "1" {
			t.Errorf("%s: missing X-Admin header", path)
		}
	}
	if rec := do(t, e, http.MethodGet, "/public"); rec.Header().Get("X-Admin") != "" {
		t.Error("/public: unexpected X-Admin header")
	}
}

func TestRegister_MiddlewareOrder(t *testing.T) {
	e := echo.New()
	var order []string
	mw := func(label string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				order = append(order, label)
				return next(c)
			}
		}
	}
	route.Register(e,
		route.Layout(
			route.Use(mw("first")),
			route.Use(mw("second")),
			route.GET("/ordered", "ordered", ok),
		),
	)
	do(t, e, http.MethodGet, "/ordered")
	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Errorf("middleware order = %v, want [first second]", order)
	}
}

func TestRegister_MixedSiblings(t *testing.T) {
	// Routes, Groups, and Layouts as siblings all receive the scope middleware.
	e := echo.New()
	route.Register(e,
		route.Layout(
			route.Use(headerMW("X-Scope", "1")),
			route.GET("/a", "a", ok),
			route.Group("/grp",
				route.GET("/b", "b", ok),
			),
			route.Layout(
				route.GET("/c", "c", ok),
			),
		),
	)

	for _, path := range []string{"/a", "/grp/b", "/c"} {
		rec := do(t, e, http.MethodGet, path)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want 200", path, rec.Code)
		}
		if rec.Header().Get("X-Scope") != "1" {
			t.Errorf("%s: missing X-Scope header", path)
		}
	}
}
