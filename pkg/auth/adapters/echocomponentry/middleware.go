package echocomponentry

import (
	"errors"
	"net/http"

	"github.com/go-sum/auth/model"
	authrepo "github.com/go-sum/auth/repository"
	"github.com/go-sum/auth/session"
	componenthtmx "github.com/go-sum/componentry/patterns/htmx"
	"github.com/go-sum/server/apperr"
	"github.com/go-sum/server/ctxkeys"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// LoadSession reads the session non-destructively and sets ctxkeys.UserID in context.
func LoadSession(sessions *session.SessionManager) echo.MiddlewareFunc {
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
func RequireAuth(loginPath string) echo.MiddlewareFunc {
	return RequireAuthPath(func() string { return loginPath })
}

// RequireAuthPath protects routes by checking for a valid session user ID.
func RequireAuthPath(loginPath func() string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userID, _ := c.Get(string(ctxkeys.UserID)).(string)
			if userID == "" {
				path := loginPath()
				if componenthtmx.NewRequest(c.Request()).IsPartial() {
					componenthtmx.Response{Redirect: path}.Apply(c.Response())
					return c.NoContent(http.StatusUnauthorized)
				}
				return c.Redirect(http.StatusSeeOther, path)
			}
			return next(c)
		}
	}
}

// LoadUserContext resolves the authenticated user's role and display name.
func LoadUserContext(users authrepo.UserReader) echo.MiddlewareFunc {
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
			c.Set(string(ctxkeys.UserDisplayName), user.DisplayName)
			return next(c)
		}
	}
}

// RequireAdmin ensures the authenticated user has the admin role.
func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role, _ := c.Get(string(ctxkeys.UserRole)).(string)
			if role != model.RoleAdmin {
				return apperr.Forbidden("You are not allowed to access this area.")
			}
			return next(c)
		}
	}
}
