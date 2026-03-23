package echocomponentry

import (
	"errors"
	"net/http"

	"github.com/go-sum/auth/model"
	authrepo "github.com/go-sum/auth/repository"
	"github.com/go-sum/auth/session"
	htmx "github.com/go-sum/componentry/patterns/htmx"
	"github.com/go-sum/server/apperr"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// ContextKeys holds the Echo context key names used by auth middleware.
// Values come from the application's ContextKeysConfig at wiring time.
type ContextKeys struct {
	UserID      string
	UserRole    string
	DisplayName string
}

// LoadSession reads the session non-destructively and sets the user ID in context.
func LoadSession(sessions *session.SessionManager, keys ContextKeys) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if userID, err := sessions.GetUserID(c.Request()); err == nil && userID != "" {
				c.Set(keys.UserID, userID)
			}
			return next(c)
		}
	}
}

// RequireAuth protects routes by checking for a valid session user ID.
func RequireAuth(signinPath string, keys ContextKeys) echo.MiddlewareFunc {
	return RequireAuthPath(func() string { return signinPath }, keys)
}

// RequireAuthPath protects routes by checking for a valid session user ID.
func RequireAuthPath(signinPath func() string, keys ContextKeys) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userID, _ := c.Get(keys.UserID).(string)
			if userID == "" {
				path := signinPath()
				if htmx.NewRequest(c.Request()).IsPartial() {
					htmx.Response{Redirect: path}.Apply(c.Response())
					return c.NoContent(http.StatusUnauthorized)
				}
				return c.Redirect(http.StatusSeeOther, path)
			}
			return next(c)
		}
	}
}

// LoadUserContext resolves the authenticated user's role and display name.
func LoadUserContext(users authrepo.UserReader, keys ContextKeys) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userID, _ := c.Get(keys.UserID).(string)
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

			c.Set(keys.UserRole, user.Role)
			c.Set(keys.DisplayName, user.DisplayName)
			return next(c)
		}
	}
}

// RequireAdmin ensures the authenticated user has the admin role.
func RequireAdmin(keys ContextKeys) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role, _ := c.Get(keys.UserRole).(string)
			if role != model.RoleAdmin {
				return apperr.Forbidden("You are not allowed to access this area.")
			}
			return next(c)
		}
	}
}
