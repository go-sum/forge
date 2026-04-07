package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/server/route"
	"github.com/labstack/echo/v5"
)

func TestBuildGroups_TopLevel(t *testing.T) {
	e := echo.New()
	groups := route.BuildGroups(e, []route.GroupDef{
		{Name: "public", Prefix: "/pub"},
	})
	if groups["public"] == nil {
		t.Fatal("expected public group, got nil")
	}
}

func TestBuildGroups_ParentChild(t *testing.T) {
	e := echo.New()
	var childCalled bool
	groups := route.BuildGroups(e, []route.GroupDef{
		{Name: "parent", Prefix: "/parent"},
		{Name: "child", Prefix: "/child", Parent: "parent"},
	})
	noOp := func(c *echo.Context) error { childCalled = true; return nil }
	route.Add(groups["child"], echo.Route{Method: http.MethodGet, Path: "", Handler: noOp})

	req := httptest.NewRequest(http.MethodGet, "/parent/child", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if !childCalled {
		t.Fatal("child group handler was not called")
	}
}

func TestBuildGroups_MiddlewareApplied(t *testing.T) {
	e := echo.New()
	var mwCalled bool
	mw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			mwCalled = true
			return next(c)
		}
	}
	groups := route.BuildGroups(e, []route.GroupDef{
		{Name: "g", Prefix: "", Middleware: []echo.MiddlewareFunc{mw}},
	})
	route.Add(groups["g"], echo.Route{Method: http.MethodGet, Path: "/ping", Handler: func(c *echo.Context) error { return nil }})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if !mwCalled {
		t.Fatal("middleware was not called")
	}
}

func TestBuildGroups_Empty(t *testing.T) {
	e := echo.New()
	groups := route.BuildGroups(e, nil)
	if len(groups) != 0 {
		t.Fatalf("expected empty map, got len %d", len(groups))
	}
}

func TestBuildGroups_PanicDuplicateName(t *testing.T) {
	e := echo.New()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for duplicate name, got none")
		}
	}()
	route.BuildGroups(e, []route.GroupDef{
		{Name: "dup", Prefix: ""},
		{Name: "dup", Prefix: "/other"},
	})
}

func TestBuildGroups_PanicUnknownParent(t *testing.T) {
	e := echo.New()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unknown parent, got none")
		}
	}()
	route.BuildGroups(e, []route.GroupDef{
		{Name: "orphan", Prefix: "", Parent: "nonexistent"},
	})
}
