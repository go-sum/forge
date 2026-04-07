package route

import "github.com/labstack/echo/v5"

// Resolver provides deferred route path resolution. It holds a lazy
// func() echo.Routes supplier so that path lookups are deferred until
// after all routes have been registered.
type Resolver struct {
	routes func() echo.Routes
}

// NewResolver creates a Resolver that calls routes() on each resolution
// to obtain the current route table.
func NewResolver(routes func() echo.Routes) *Resolver {
	return &Resolver{routes: routes}
}

// Routes returns the underlying lazy route supplier. Use when a callee
// needs func() echo.Routes directly (e.g. sitemap generation).
func (r *Resolver) Routes() func() echo.Routes {
	return r.routes
}

// Path returns a func() string that resolves the named route to its path.
// Panics if the route name is unknown (consistent with route.Reverse).
func (r *Resolver) Path(name string, pathValues ...any) func() string {
	return func() string {
		return Reverse(r.routes(), name, pathValues...)
	}
}

// URL returns a func() string that resolves the named route and prepends
// origin (scheme + host). Use for absolute URLs such as email verification links.
func (r *Resolver) URL(origin, name string, pathValues ...any) func() string {
	return func() string {
		return origin + Reverse(r.routes(), name, pathValues...)
	}
}
