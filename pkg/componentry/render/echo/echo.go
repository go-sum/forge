// Package echo provides Echo v5 adapters for componentry/render.
//
// Component and Fragment are the two primary entry points. Both perform the same
// rendering operation — the distinction is semantic and signals caller intent:
//   - Component: conventionally used for full-page renders
//   - Fragment: conventionally used for HTMX partial renders
//
// Both have WithStatus variants for non-200 responses (e.g. 422 on validation errors).
package echo

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/go-sum/componentry/render"
	g "maragu.dev/gomponents"
)

// Component renders a gomponents node as a full HTML response with 200 OK.
func Component(c *echo.Context, node g.Node) error {
	return ComponentWithStatus(c, http.StatusOK, node)
}

// ComponentWithStatus renders a gomponents node with a custom HTTP status code.
func ComponentWithStatus(c *echo.Context, status int, node g.Node) error {
	return render.ComponentWithStatus(c.Response(), status, node)
}

// Fragment renders a gomponents node as an HTMX partial response with 200 OK.
func Fragment(c *echo.Context, node g.Node) error {
	return ComponentWithStatus(c, http.StatusOK, node)
}

// FragmentWithStatus renders a partial with a custom HTTP status code.
func FragmentWithStatus(c *echo.Context, status int, node g.Node) error {
	return ComponentWithStatus(c, status, node)
}
