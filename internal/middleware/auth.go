package middleware

import (
	"errors"
	"net/http"

	"starter/internal/apperr"
	"starter/internal/model"
	"starter/internal/service"
	"starter/pkg/auth"
	componenthtmx "starter/pkg/components/patterns/htmx"
	"starter/pkg/ctxkeys"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// htmxRedirectHeader is the HTMX response header that triggers a client-side redirect.
const htmxRedirectHeader = "HX-Redirect"

// LoadSession reads the session non-destructively and sets ctxkeys.UserID in context
// when a valid session exists. Unlike RequireAuth, it never redirects unauthenticated
// requests — it is safe to apply globally so every handler can inspect auth state.
func LoadSession(sessions *auth.SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if userID, err := sessions.GetUserID(c.Request()); err == nil && userID != "" {
				c.Set(string(ctxkeys.UserID), userID)
			}
			return next(c)
		}
	}
}

// RequireAuth protects routes by checking for a valid session user ID.
// It reads ctxkeys.UserID set by LoadSession — LoadSession must run first.
// loginPath is the URL unauthenticated requests are redirected to.
// HTMX requests receive HX-Redirect instead of a standard 3xx redirect.
func RequireAuth(loginPath string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userID, _ := c.Get(string(ctxkeys.UserID)).(string)
			if userID == "" {
				if componenthtmx.IsRequest(c.Request()) {
					c.Response().Header().Set(htmxRedirectHeader, loginPath)
					return c.NoContent(http.StatusUnauthorized)
				}
				return c.Redirect(http.StatusSeeOther, loginPath)
			}
			return next(c)
		}
	}
}

// LoadUserRole resolves the authenticated user's role and stores it in context.
func LoadUserRole(users *service.UserService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userID, _ := c.Get(string(ctxkeys.UserID)).(string)
			if userID == "" {
				return next(c)
			}

			id, err := uuid.Parse(userID)
			if err != nil {
				return apperr.Unauthorized("Your session is invalid. Please sign in again.")
			}

			user, err := users.GetByID(c.Request().Context(), id)
			if err != nil {
				if errors.Is(err, model.ErrUserNotFound) {
					return apperr.Unauthorized("Your account could not be loaded. Please sign in again.")
				}
				return apperr.Unavailable("Unable to authorize this request right now.", err)
			}

			c.Set(string(ctxkeys.UserRole), user.Role)
			return next(c)
		}
	}
}

// RequireAdmin ensures the authenticated user has the admin role.
func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role, _ := c.Get(string(ctxkeys.UserRole)).(string)
			if role != "admin" {
				return apperr.Forbidden("You are not allowed to access this area.")
			}
			return next(c)
		}
	}
}
