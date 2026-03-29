// Package override provides an HTTP method override middleware for Echo v5.
//
// HTML forms only emit GET and POST. This middleware reads a configurable form
// field (default "_method") from POST request bodies and promotes the request
// method to the specified value before routing. Only PUT, PATCH, and DELETE are
// permitted override targets — any other value results in a 400 Bad Request.
//
// The middleware is designed for e.Pre() registration so the promoted method is
// visible to the router before dispatch.
package override

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

// allowedOverrideMethods is the closed set of methods that may be promoted via
// form field override. GET, HEAD, OPTIONS, and TRACE are never valid targets.
var allowedOverrideMethods = map[string]struct{}{
	http.MethodPut:    {},
	http.MethodPatch:  {},
	http.MethodDelete: {},
}

// Config defines the method override middleware configuration.
type Config struct {
	// Skipper defines a function to skip the middleware. Defaults to never skip.
	Skipper func(c *echo.Context) bool

	// FormField is the POST body field name that carries the override verb.
	// Defaults to "_method".
	FormField string
}

// DefaultConfig is the default Config used by Middleware().
var DefaultConfig = Config{
	FormField: "_method",
}

// Middleware returns an Echo middleware with DefaultConfig.
func Middleware() echo.MiddlewareFunc {
	return NewWithConfig(DefaultConfig)
}

// NewWithConfig returns an Echo middleware with the given config.
// Panics if the config is invalid.
func NewWithConfig(cfg Config) echo.MiddlewareFunc {
	mw, err := cfg.ToMiddleware()
	if err != nil {
		panic(err)
	}
	return mw
}

// ToMiddleware converts Config to an echo.MiddlewareFunc or returns an error
// for invalid configuration. Satisfies the echo.MiddlewareConfigurator interface.
func (cfg Config) ToMiddleware() (echo.MiddlewareFunc, error) {
	if cfg.Skipper == nil {
		cfg.Skipper = func(*echo.Context) bool { return false }
	}
	if cfg.FormField == "" {
		cfg.FormField = DefaultConfig.FormField
	}

	field := cfg.FormField

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}

			// Only POST requests carry a form body that may contain the override.
			if c.Request().Method != http.MethodPost {
				return next(c)
			}

			raw := c.FormValue(field)
			if raw == "" {
				return next(c)
			}

			method := strings.ToUpper(raw)
			if _, ok := allowedOverrideMethods[method]; !ok {
				return &violation{
					status:  http.StatusBadRequest,
					message: "Method override value is not permitted.",
				}
			}

			c.Request().Method = method
			return next(c)
		}
	}, nil
}

// violation is a typed method-override failure. It implements StatusCode() and
// PublicMessage() so that the application's error-classification pipeline emits
// the correct HTTP status (400 Bad Request) rather than falling through to 500.
type violation struct {
	status  int
	message string
}

func (v *violation) Error() string         { return v.message }
func (v *violation) StatusCode() int       { return v.status }
func (v *violation) PublicMessage() string { return v.message }
