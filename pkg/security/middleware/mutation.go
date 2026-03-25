package middleware

import (
	"net/http"

	"github.com/go-sum/security/fetchmeta"
	"github.com/go-sum/security/httpsec"
	"github.com/go-sum/security/origin"
	"github.com/labstack/echo/v5"
)

// CrossOriginGuard applies origin and Fetch Metadata checks to unsafe requests.
func CrossOriginGuard(originPolicy origin.Policy, fetchPolicy fetchmeta.Policy) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			if !httpsec.IsUnsafeMethod(req.Method) {
				return next(c)
			}

			if result := origin.Validate(req, originPolicy); !result.Valid {
				return &Error{
					Status:  http.StatusForbidden,
					Message: originFailureMessage(result),
				}
			}
			if result := fetchmeta.Validate(req, fetchPolicy); !result.Valid {
				return &Error{
					Status:  http.StatusForbidden,
					Message: fetchMetadataFailureMessage(result),
				}
			}

			return next(c)
		}
	}
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
