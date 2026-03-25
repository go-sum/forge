package route

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/labstack/echo/v5"
)

type RouteAdder interface {
	AddRoute(route echo.Route) (echo.RouteInfo, error)
}

func Add(target RouteAdder, route echo.Route) echo.RouteInfo {
	info, err := target.AddRoute(route)
	if err != nil {
		panic(fmt.Sprintf("route %s %s (%s): %v", route.Method, route.Path, route.Name, err))
	}
	return info
}

func Reverse(routes echo.Routes, name string, pathValues ...any) string {
	path, err := routes.Reverse(name, pathValues...)
	if err != nil {
		panic(fmt.Sprintf("reverse %q: %v", name, err))
	}
	return path
}

func ReverseWithQuery(routes echo.Routes, name string, query url.Values, pathValues ...any) string {
	path := Reverse(routes, name, pathValues...)
	if len(query) == 0 {
		return path
	}

	encoded := query.Encode()
	if encoded == "" {
		return path
	}
	return path + "?" + encoded
}

// SafeReverse resolves a named route to its path without panicking.
// Returns ("", false) if the route name is unknown or the resolved path
// still contains ":" — indicating an unfilled path parameter (e.g.
// /users/:id/edit). Such routes produce invalid sitemap URLs and are skipped.
func SafeReverse(routes echo.Routes, name string) (string, bool) {
	path, err := routes.Reverse(name)
	if err != nil {
		return "", false
	}
	if strings.Contains(path, ":") {
		return "", false
	}
	return path, true
}
