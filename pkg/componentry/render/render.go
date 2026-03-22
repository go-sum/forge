// Package render provides helpers for writing gomponents nodes as HTTP responses.
//
// This package uses only stdlib net/http. For Echo-specific helpers see the
// render/echo sub-package.
package render

import (
	"net/http"

	g "maragu.dev/gomponents"
)

const mimeTextHTMLCharsetUTF8 = "text/html; charset=UTF-8"

// Component renders a gomponents node as a full HTML response with 200 OK.
func Component(w http.ResponseWriter, node g.Node) error {
	return ComponentWithStatus(w, http.StatusOK, node)
}

// ComponentWithStatus renders a gomponents node with a custom HTTP status code.
func ComponentWithStatus(w http.ResponseWriter, status int, node g.Node) error {
	w.Header().Set("Content-Type", mimeTextHTMLCharsetUTF8)
	w.WriteHeader(status)
	return node.Render(w)
}

// Fragment renders a gomponents node as an HTMX partial response with 200 OK.
func Fragment(w http.ResponseWriter, node g.Node) error {
	return ComponentWithStatus(w, http.StatusOK, node)
}

// FragmentWithStatus renders a partial with a custom HTTP status code.
func FragmentWithStatus(w http.ResponseWriter, status int, node g.Node) error {
	return ComponentWithStatus(w, status, node)
}
