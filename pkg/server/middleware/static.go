// Package middleware provides reusable Echo middleware functions.
// Each file contains a single middleware or a closely related group.
// These are generic — they carry no application-specific imports.
package middleware

import (
	"strings"

	"github.com/labstack/echo/v5"
)

// cacheImmutable is the Cache-Control value for versioned static assets: cached for one year,
// never revalidated. Safe because asset URLs include a content hash (?v=<hash>).
const cacheImmutable = "public, max-age=31536000, immutable"

// StaticCacheControl returns middleware that sets Cache-Control headers for
// static asset paths under prefix. Versioned assets (with ?v=<hash>) receive
// a one-year immutable header; unversioned assets get no-cache.
//
// Accepting prefix as a parameter eliminates the middleware silently doing nothing if the URL prefix changes.
func StaticCacheControl(prefix string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if prefix != "" && strings.HasPrefix(c.Request().URL.Path, prefix+"/") {
				if c.Request().URL.Query().Get("v") != "" {
					c.Response().Header().Set("Cache-Control", cacheImmutable)
				} else {
					c.Response().Header().Set("Cache-Control", "no-cache")
				}
			}
			return next(c)
		}
	}
}
