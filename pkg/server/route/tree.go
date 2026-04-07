package route

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// nodeKind identifies which variant a Node represents.
type nodeKind int

const (
	nodeRoute  nodeKind = iota // leaf endpoint
	nodeGroup                  // URL prefix + children
	nodeLayout                 // middleware scope, no URL change
	nodeUse                    // middleware declaration
)

// Node is an element in a route tree definition.
// Construct nodes with Route, Group, Layout, and Use —
// never build the struct directly.
type Node struct {
	kind       nodeKind
	method     string
	path       string
	name       string
	handler    echo.HandlerFunc
	middleware []echo.MiddlewareFunc
	children   []Node
}

// Route defines a leaf endpoint.
func Route(method, path, name string, h echo.HandlerFunc) Node {
	return Node{kind: nodeRoute, method: method, path: path, name: name, handler: h}
}

// GET is a shorthand for Route(http.MethodGet, ...).
func GET(path, name string, h echo.HandlerFunc) Node {
	return Route(http.MethodGet, path, name, h)
}

// POST is a shorthand for Route(http.MethodPost, ...).
func POST(path, name string, h echo.HandlerFunc) Node {
	return Route(http.MethodPost, path, name, h)
}

// PUT is a shorthand for Route(http.MethodPut, ...).
func PUT(path, name string, h echo.HandlerFunc) Node {
	return Route(http.MethodPut, path, name, h)
}

// DELETE is a shorthand for Route(http.MethodDelete, ...).
func DELETE(path, name string, h echo.HandlerFunc) Node {
	return Route(http.MethodDelete, path, name, h)
}

// Group nests children under a URL prefix. Any Use() nodes inside
// children apply as middleware to all routes in the group.
func Group(prefix string, children ...Node) Node {
	return Node{kind: nodeGroup, path: prefix, children: children}
}

// Layout nests children under shared middleware without adding a URL
// prefix. Use() nodes inside children declare that middleware.
// A Layout with no Use() nodes is a structural grouping with no
// runtime cost — no extra echo.Group is created.
func Layout(children ...Node) Node {
	return Node{kind: nodeLayout, children: children}
}

// Use declares middleware for the enclosing Group or Layout.
// All Use() nodes in a scope are collected before routes are processed;
// their order relative to Route/Group/Layout siblings does not matter,
// but order among multiple Use() calls is preserved.
func Use(mw ...echo.MiddlewareFunc) Node {
	return Node{kind: nodeUse, middleware: mw}
}

// Register walks the route tree and registers all routes on e.
// It panics on registration errors, consistent with Add and Reverse.
func Register(e *echo.Echo, nodes ...Node) {
	walk(e, nodes)
}

// routeTarget is satisfied by both *echo.Echo and *echo.Group.
type routeTarget interface {
	AddRoute(echo.Route) (echo.RouteInfo, error)
	Group(prefix string, m ...echo.MiddlewareFunc) *echo.Group
}

func walk(t routeTarget, nodes []Node) {
	// Pass 1: collect middleware from all Use() nodes in this scope.
	var mw []echo.MiddlewareFunc
	for _, n := range nodes {
		if n.kind == nodeUse {
			mw = append(mw, n.middleware...)
		}
	}

	// If middleware was declared, create a scoped echo.Group to hold it.
	// All routes and sub-groups in this scope are registered on that group,
	// inheriting the middleware via Echo's parent-snapshot mechanism.
	target := t
	if len(mw) > 0 {
		target = t.Group("", mw...)
	}

	// Pass 2: process Route, Group, and Layout nodes.
	for _, n := range nodes {
		switch n.kind {
		case nodeRoute:
			Add(target, echo.Route{
				Method:  n.method,
				Path:    n.path,
				Name:    n.name,
				Handler: n.handler,
			})
		case nodeGroup:
			g := target.Group(n.path)
			walk(g, n.children)
		case nodeLayout:
			walk(target, n.children)
		case nodeUse:
			// handled in pass 1
		}
	}
}
