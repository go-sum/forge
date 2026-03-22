// Package redirect provides a builder for issuing HTTP redirects that are
// HTMX-aware. It depends only on net/http — no framework import.
// HX-Request detection is done inline to preserve the leaf-node boundary.
package redirect

import (
	"net/http"
)

// Builder constructs and executes a redirect response.
type Builder struct {
	w      http.ResponseWriter
	r      *http.Request
	url    string
	status int
}

// New starts a redirect builder for the given request/response pair.
func New(w http.ResponseWriter, r *http.Request) *Builder {
	return &Builder{w: w, r: r, status: http.StatusSeeOther}
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

// Go executes the redirect. For HTMX requests it sets HX-Redirect and responds
// with 204 No Content so the client-side swap can follow the redirect. For
// boosted requests and plain HTTP requests a standard 3xx redirect is issued.
func (b *Builder) Go() error {
	if b.r.Header.Get("HX-Request") == "true" && b.r.Header.Get("HX-Boosted") != "true" {
		b.w.Header().Set("HX-Redirect", b.url)
		b.w.WriteHeader(http.StatusNoContent)
		return nil
	}
	http.Redirect(b.w, b.r, b.url, b.status)
	return nil
}
