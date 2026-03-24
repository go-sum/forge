// Package ratelimit wraps Echo's built-in rate limiter middleware behind a
// Config type and ensures that all failure paths return typed errors compatible
// with the application's error-classification pipeline.
//
// Echo's default DenyHandler and ErrorHandler return raw *echo.HTTPError values
// that do not implement StatusCode() int or PublicMessage() string. The app's
// classify() function in internal/server/error.go would therefore fall through
// to apperr.Internal() and produce a 500 instead of a 429 Too Many Requests.
// This package installs typed handlers that fix that classification.
package ratelimit

import (
	"log/slog"
	"net"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// Config defines the rate limiting policy.
type Config struct {
	// Rate is the number of requests allowed per second (token bucket refill).
	Rate float64

	// Burst is the maximum number of requests allowed above the steady-state
	// rate at any instant. Defaults to ceil(Rate) when zero.
	Burst int

	// IdentifierExtractor extracts the rate limit key from the request context.
	// Defaults to the remote IP address when nil.
	IdentifierExtractor func(c *echo.Context) (string, error)
}

// violation is a typed rate-limit failure value. It implements StatusCode() and
// PublicMessage() so that internal/server.classify() emits the correct HTTP
// status response rather than falling through to a 500 Internal Server Error.
type violation struct {
	code int
	msg  string
}

func (v *violation) Error() string         { return v.msg }
func (v *violation) StatusCode() int       { return v.code }
func (v *violation) PublicMessage() string { return v.msg }

const denyMessage = "Too many requests. Please wait and try again."

// Middleware returns an Echo middleware that enforces a per-identifier request
// rate using Echo's built-in in-memory token bucket store.
//
// Requests exceeding the configured rate return a typed 429 Too Many Requests
func Middleware(cfg Config) echo.MiddlewareFunc {
	store := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:  cfg.Rate,
			Burst: cfg.Burst,
		},
	)

	extractor := cfg.IdentifierExtractor
	if extractor == nil {
		extractor = func(c *echo.Context) (string, error) {
			// c.RealIP() reads X-Forwarded-For as-is, which may be non-standard "IP:port"
			ip := c.RealIP()
			if host, _, err := net.SplitHostPort(ip); err == nil {
				return host, nil
			}
			return ip, nil
		}
	}

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store:               store,
		IdentifierExtractor: extractor,

		// DenyHandler is called when the rate limit is exceeded.
		DenyHandler: func(c *echo.Context, identifier string, _ error) error {
			slog.Default().Debug("rate limit exceeded",
				"ip", identifier,
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
			)
			return &violation{code: http.StatusTooManyRequests, msg: denyMessage}
		},

		// ErrorHandler is called when IdentifierExtractor returns an error.
		// Should may happen with with custom extractors; treat it as an internal error
		// so the request is not silently allowed through.
		ErrorHandler: func(_ *echo.Context, err error) error {
			return &violation{code: http.StatusInternalServerError, msg: err.Error()}
		},
	})
}
