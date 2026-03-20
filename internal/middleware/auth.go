package middleware

import (
	"net/http"

	"starter/pkg/auth"
	"starter/pkg/ctxkeys"

	"github.com/labstack/echo/v5"
)

// RequireAuth protects routes by checking for a valid session user ID.
// HTMX requests receive HX-Redirect instead of a standard 3xx redirect.
func RequireAuth(sessions *auth.SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userID, err := sessions.GetUserID(c.Request())
			if err != nil || userID == "" {
				if c.Request().Header.Get("HX-Request") == "true" {
					c.Response().Header().Set("HX-Redirect", "/login")
					return c.NoContent(http.StatusUnauthorized)
				}
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			c.Set(string(ctxkeys.UserID), userID)
			return next(c)
		}
	}
}
