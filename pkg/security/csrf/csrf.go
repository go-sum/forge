// Package csrf provides HMAC-signed, time-limited CSRF protection for Echo v5.
//
// Unlike the double-submit cookie approach (where a random token lives in a
// cookie and must match the submitted form field), this package issues tokens
// whose integrity is guaranteed by HMAC-SHA256. A valid token can only be
// produced by a server that holds the signing key, so no cookie comparison is
// required. Tokens carry their own expiry and cannot be replayed after they
// expire.
//
// On every safe request (GET, HEAD, OPTIONS, TRACE) the middleware issues a
// fresh token and stores it in the Echo context under Config.ContextKey. View
// templates embed the token in forms via the hidden _csrf field. On unsafe
// requests (POST, PUT, PATCH, DELETE) the submitted token is verified; any
// failure returns a typed *violation error so the application's error handler
// produces a 403 Forbidden response.
package csrf

import (
	"net/http"
	"time"

	"github.com/go-sum/security/httpsec"
	"github.com/go-sum/security/token"
	"github.com/labstack/echo/v5"
)

const (
	scope             = "csrf"
	defaultTTLSeconds = int(time.Hour / time.Second)
	failureMessage    = "Your security token is invalid or missing. Refresh the page and try again."
)

// Config defines the signing key, token lifetime, and transport field names.
type Config struct {
	// Key is the HMAC-SHA256 signing key. Must be at least 32 bytes.
	Key []byte
	// TokenTTL is how long an issued token remains valid, in seconds. Defaults to 3600.
	TokenTTL int
	// ContextKey is the Echo context key under which the token string is stored.
	ContextKey string
	// HeaderName is the request header checked before FormField on unsafe methods.
	HeaderName string
	// FormField is the form field name read when HeaderName is absent.
	FormField string
}

// violation is a typed CSRF failure value. It implements StatusCode() and
// PublicMessage() so that internal/server.classify() emits a 403 Forbidden
// response rather than a 500 Internal Server Error.
type violation struct{ msg string }

func (v *violation) Error() string         { return v.msg }
func (v *violation) StatusCode() int       { return http.StatusForbidden }
func (v *violation) PublicMessage() string { return v.msg }

// Middleware returns an Echo middleware that applies HMAC-signed CSRF protection.
//
// Safe methods receive a freshly issued token stored in the Echo context under
// Config.ContextKey. View templates embed this value in a hidden form field.
//
// Unsafe methods read the submitted token from Config.HeaderName first, then
// Config.FormField, and verify it with HMAC-SHA256. Any failure — missing,
// malformed, tampered, or expired token — returns a *violation error (403).
func Middleware(cfg Config) echo.MiddlewareFunc {
	ttlSeconds := cfg.TokenTTL
	if ttlSeconds <= 0 {
		ttlSeconds = defaultTTLSeconds
	}
	ttl := time.Duration(ttlSeconds) * time.Second

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if httpsec.IsSafeMethod(c.Request().Method) {
				tok, err := token.Issue(cfg.Key, scope, ttl)
				if err != nil {
					return err // crypto/rand failure; propagates as 500
				}
				c.Set(cfg.ContextKey, tok)
				return next(c)
			}

			// Unsafe method: verify the submitted token.
			raw := c.Request().Header.Get(cfg.HeaderName)
			if raw == "" {
				raw = c.FormValue(cfg.FormField)
			}
			if raw == "" {
				return &violation{failureMessage}
			}
			if err := token.Verify(cfg.Key, scope, raw); err != nil {
				return &violation{failureMessage}
			}

			// Issue a fresh token so handlers that re-render forms on POST
			// (e.g. validation failures) embed a valid token in the response.
			fresh, err := token.Issue(cfg.Key, scope, ttl)
			if err != nil {
				return err
			}
			c.Set(cfg.ContextKey, fresh)

			return next(c)
		}
	}
}
