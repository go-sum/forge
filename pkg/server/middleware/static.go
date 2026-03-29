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

// StaticCacheConfig defines cache-control behavior for static asset responses.
type StaticCacheConfig struct {
	// Skipper defines a function to skip the middleware. Defaults to never skip.
	Skipper func(c *echo.Context) bool

	// VersionParam is the query parameter name used to detect versioned assets.
	// Defaults to "v".
	VersionParam string

	// VersionedHeader is the Cache-Control value for versioned assets.
	// Defaults to "public, max-age=31536000, immutable".
	VersionedHeader string

	// UnversionedHeader is the Cache-Control value for unversioned assets.
	// Defaults to "no-cache".
	UnversionedHeader string
}

// ToMiddleware converts the config into an Echo middleware function.
func (cfg StaticCacheConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if cfg.Skipper == nil {
		cfg.Skipper = func(*echo.Context) bool { return false }
	}
	if cfg.VersionParam == "" {
		cfg.VersionParam = "v"
	}
	if cfg.VersionedHeader == "" {
		cfg.VersionedHeader = cacheImmutable
	}
	if cfg.UnversionedHeader == "" {
		cfg.UnversionedHeader = "no-cache"
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}
			if c.Request().URL.Query().Get(cfg.VersionParam) != "" {
				c.Response().Header().Set("Cache-Control", cfg.VersionedHeader)
			} else {
				c.Response().Header().Set("Cache-Control", cfg.UnversionedHeader)
			}
			return next(c)
		}
	}, nil
}

// StaticCache returns an Echo middleware that sets Cache-Control headers for
// static assets. Versioned assets (with a version query parameter) receive
// immutable caching; unversioned assets receive no-cache. Panics on invalid config.
func StaticCache(cfg StaticCacheConfig) echo.MiddlewareFunc {
	mw, err := cfg.ToMiddleware()
	if err != nil {
		panic(err)
	}
	return mw
}

// StaticCacheControl returns middleware that sets Cache-Control headers for
// static asset paths under prefix. Versioned assets (with ?v=<hash>) receive
// a one-year immutable header; unversioned assets get no-cache.
//
// Deprecated: Apply StaticCache(StaticCacheConfig{}) to a route group instead
// of using global path-prefix matching. This function remains for backward
// compatibility.
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
