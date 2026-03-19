// Package redirect provides a builder for issuing HTTP redirects that are
// HTMX-aware. HX-Request detection uses inline header reads rather than
// importing pkg/htmx to preserve the leaf-node boundary.
package redirect

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// Builder constructs and executes a redirect response.
type Builder struct {
	c      *echo.Context
	url    string
	status int
}

// New starts a redirect builder for the given Echo context.
func New(c *echo.Context) *Builder {
	return &Builder{c: c, status: http.StatusSeeOther}
}

// To sets the destination URL.
func (b *Builder) To(url string) *Builder {
	b.url = url
	return b
}

// StatusCode overrides the default 303 redirect status.
func (b *Builder) StatusCode(code int) *Builder {
	b.status = code
	return b
}

// Go executes the redirect. For HTMX requests it sets HX-Redirect instead of
// issuing a 3xx response, so the client-side swap can follow the redirect.
func (b *Builder) Go() error {
	r := (*b.c).Request()
	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true" {
		(*b.c).Response().Header().Set("HX-Redirect", b.url)
		return (*b.c).NoContent(http.StatusNoContent)
	}
	return (*b.c).Redirect(b.status, b.url)
}
