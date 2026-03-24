package middleware

import (
	"github.com/go-sum/security/headers"
	"github.com/labstack/echo/v5"
)

// SecurityHeaders applies reusable security headers to every response.
func SecurityHeaders(policy headers.Policy) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			headers.Apply(c.Response().Header(), policy)
			return next(c)
		}
	}
}
