package route

import (
	"fmt"
	"net/url"

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
