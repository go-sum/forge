package middleware

import (
	"github.com/go-sum/server/headers"
	"github.com/labstack/echo/v5"
)

// CacheHeaders returns middleware that sets Cache-Control and appends Vary
// values before calling the next handler.
func CacheHeaders(cacheControl string, vary ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if cacheControl != "" {
				c.Response().Header().Set("Cache-Control", cacheControl)
			}
			headers.AppendVary(c.Response().Header(), vary...)
			return next(c)
		}
	}
}
