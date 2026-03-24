// Package csrf wraps Echo's built-in double-submit cookie CSRF middleware behind
// a transport-neutral Config type and ensures that all failure paths return typed
// errors compatible with the application's error-classification pipeline.
//
// Echo v5's CSRF middleware uses Sec-Fetch-Site as the primary defence for modern
// browsers and falls back to cookie-plus-token comparison for older ones.
// Two error paths exist in the underlying middleware:
//
//  1. Token mismatch — goes through CSRFConfig.ErrorHandler.
//  2. Sec-Fetch-Site cross-site/same-site block — goes through
//     CSRFConfig.AllowSecFetchSiteFunc and returns an error directly.
//
// Both paths return a *violation value so that the app's classify() function
// can detect StatusCode() and PublicMessage() and produce a 403 Forbidden
// response rather than falling through to a 500 Internal Server Error.
package csrf

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// Config defines CSRF cookie and token lookup policy.
type Config struct {
	ContextKey   string
	HeaderName   string
	FormField    string
	CookieName   string
	CookieSecure bool
}

// violation is a typed CSRF failure value. It implements StatusCode() and
// PublicMessage() so that internal/server.classify() emits a 403 Forbidden
// response rather than a 500 Internal Server Error.
type violation struct{ msg string }

func (v *violation) Error() string         { return v.msg }
func (v *violation) StatusCode() int       { return http.StatusForbidden }
func (v *violation) PublicMessage() string { return v.msg }

const failureMessage = "Your security token is invalid or missing. Refresh the page and try again."

// Middleware returns an Echo middleware that applies CSRF protection using
// Echo's built-in double-submit cookie implementation.
//
// On safe methods (GET, HEAD, OPTIONS, TRACE) the middleware generates or
// reuses a token, writes it to the CSRF cookie, and stores the value in the
// Echo context under Config.ContextKey so that view templates can embed it
// in hidden form fields.
//
// On unsafe methods (POST, PUT, DELETE, PATCH) the submitted token — read from
// Config.HeaderName first, then Config.FormField — must match the CSRF cookie.
// Modern browsers that send Sec-Fetch-Site: same-origin bypass token comparison.
// Any failure returns a *violation error (403 Forbidden with a safe message).
func Middleware(cfg Config) echo.MiddlewareFunc {
	return middleware.CSRFWithConfig(middleware.CSRFConfig{
		ContextKey:     cfg.ContextKey,
		TokenLookup:    "header:" + cfg.HeaderName + ",form:" + cfg.FormField,
		CookieName:     cfg.CookieName,
		CookieSameSite: http.SameSiteLaxMode,
		CookieSecure:   cfg.CookieSecure,
		CookiePath:     "/",

		// ErrorHandler covers token-mismatch and missing-token failures.
		ErrorHandler: func(_ *echo.Context, _ error) error {
			return &violation{failureMessage}
		},

		// AllowSecFetchSiteFunc is called for state-changing requests where
		// Sec-Fetch-Site is "same-site" or "cross-site". Returning (false, err)
		// blocks the request; the typed error ensures correct 403 classification.
		AllowSecFetchSiteFunc: func(_ *echo.Context) (bool, error) {
			return false, &violation{failureMessage}
		},
	})
}
