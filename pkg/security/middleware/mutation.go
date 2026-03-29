package middleware

import (
	"errors"
	"net/http"

	"github.com/go-sum/security/fetchmeta"
	"github.com/go-sum/security/httpsec"
	"github.com/go-sum/security/origin"
	"github.com/labstack/echo/v5"
)

// Config defines the CrossOriginGuard middleware configuration.
type Config struct {
	// Skipper defines a function to skip the middleware. Defaults to never skip.
	Skipper func(c *echo.Context) bool

	// OriginPolicy configures origin header validation for unsafe requests.
	OriginPolicy origin.Policy

	// FetchPolicy configures Fetch Metadata validation for unsafe requests.
	FetchPolicy fetchmeta.Policy
}

// ToMiddleware converts Config to an echo.MiddlewareFunc, returning an error
// for invalid configuration rather than panicking.
func (cfg Config) ToMiddleware() (echo.MiddlewareFunc, error) {
	if cfg.Skipper == nil {
		cfg.Skipper = func(*echo.Context) bool { return false }
	}

	if !cfg.OriginPolicy.Enabled && !cfg.FetchPolicy.Enabled {
		return nil, errors.New("crossorigin: at least one of OriginPolicy or FetchPolicy must be enabled")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			if !httpsec.IsUnsafeMethod(req.Method) {
				return next(c)
			}

			if cfg.OriginPolicy.Enabled {
				if result := origin.Validate(req, cfg.OriginPolicy); !result.Valid {
					return &Error{
						Status:  http.StatusForbidden,
						Message: originFailureMessage(result),
					}
				}
			}
			if cfg.FetchPolicy.Enabled {
				if result := fetchmeta.Validate(req, cfg.FetchPolicy); !result.Valid {
					return &Error{
						Status:  http.StatusForbidden,
						Message: fetchMetadataFailureMessage(result),
					}
				}
			}

			return next(c)
		}
	}, nil
}

// Middleware returns an Echo middleware from cfg. Panics on invalid configuration.
func Middleware(cfg Config) echo.MiddlewareFunc {
	mw, err := cfg.ToMiddleware()
	if err != nil {
		panic(err)
	}
	return mw
}

// CrossOriginGuard applies origin and Fetch Metadata checks to unsafe requests.
// Deprecated: Use Config.ToMiddleware() or Middleware(Config{...}) instead.
func CrossOriginGuard(originPolicy origin.Policy, fetchPolicy fetchmeta.Policy) echo.MiddlewareFunc {
	return Middleware(Config{
		OriginPolicy: originPolicy,
		FetchPolicy:  fetchPolicy,
	})
}

func originFailureMessage(result origin.Result) string {
	switch {
	case result.HeadersMissing:
		return "This request is missing required origin headers. Refresh the page and try again."
	case result.Source != "":
		return "This request was blocked by " + result.Source + " validation. Refresh the page and try again."
	default:
		return "This request could not be verified. Refresh the page and try again."
	}
}

func fetchMetadataFailureMessage(result fetchmeta.Result) string {
	if result.HeadersMissing {
		return "This request is missing required browser security metadata."
	}
	return "This request was blocked by the browser security policy."
}
