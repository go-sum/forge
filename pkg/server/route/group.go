package route

import "github.com/labstack/echo/v5"

// GroupDef declares a named route group with an optional parent and middleware.
// Use BuildGroups to construct the groups from an ordered slice of definitions.
type GroupDef struct {
	Name       string
	Prefix     string
	Parent     string // name of a previously-defined GroupDef; empty = child of Echo
	Middleware []echo.MiddlewareFunc
}

// BuildGroups creates echo.Group instances from an ordered slice of GroupDef.
// Each definition may reference a parent by name; that parent must appear
// earlier in the slice. Top-level groups (Parent == "") are children of e.
// The returned map is keyed by GroupDef.Name.
//
// BuildGroups panics if a name is duplicated or a parent name is not found.
func BuildGroups(e *echo.Echo, defs []GroupDef) map[string]*echo.Group {
	groups := make(map[string]*echo.Group, len(defs))
	for _, d := range defs {
		if _, dup := groups[d.Name]; dup {
			panic("route.BuildGroups: duplicate group name: " + d.Name)
		}
		var g *echo.Group
		if d.Parent == "" {
			g = e.Group(d.Prefix)
		} else {
			parent, ok := groups[d.Parent]
			if !ok {
				panic("route.BuildGroups: unknown parent " + d.Parent + " for group " + d.Name)
			}
			g = parent.Group(d.Prefix)
		}
		g.Use(d.Middleware...)
		groups[d.Name] = g
	}
	return groups
}
