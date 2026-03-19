// Package render provides helpers for writing gomponents nodes as HTTP responses.
//
// Component and Fragment are the two primary entry points. Both perform the same
// rendering operation — the distinction is semantic and signals caller intent:
//   - Component: conventionally used for full-page renders
//   - Fragment: conventionally used for HTMX partial renders
//
// Both have WithStatus variants for non-200 responses (e.g. 422 on validation errors).
package render

import (
	"net/http"

	"github.com/labstack/echo/v5"
	g "maragu.dev/gomponents"
)

// Component renders a gomponents node as a full HTML response with 200 OK.
func Component(c *echo.Context, node g.Node) error {
	return ComponentWithStatus(c, http.StatusOK, node)
}

// ComponentWithStatus renders a gomponents node with a custom HTTP status code.
//
// Header.Set must precede WriteHeader — once WriteHeader is called, headers are
// flushed and subsequent Set calls are silently ignored.
func ComponentWithStatus(c *echo.Context, status int, node g.Node) error {
	w := c.Response()
	w.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	w.WriteHeader(status)
	return node.Render(w)
}

// Fragment renders a gomponents node as an HTMX partial response with 200 OK.
func Fragment(c *echo.Context, node g.Node) error {
	return ComponentWithStatus(c, http.StatusOK, node)
}

// FragmentWithStatus renders a partial with a custom HTTP status code.
// Use this for HTMX handlers that need to return non-200 status alongside
// partial HTML — e.g. status 422 with an inline validation error component.
func FragmentWithStatus(c *echo.Context, status int, node g.Node) error {
	return ComponentWithStatus(c, status, node)
}
