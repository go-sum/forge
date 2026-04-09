package auth

import (
	"errors"
	"net/http"

	"github.com/go-sum/auth/model"
	"github.com/go-sum/auth/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// LoadSession reads the session non-destructively and sets the user ID in context.
func LoadSession(mgr SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			state, err := mgr.Load(c.Request())
			if err != nil {
				return next(c)
			}
			if userID, ok := getUserID(state); ok {
				c.Set(ContextKeyUserID, userID)
				if name, ok := getDisplayName(state); ok {
					c.Set(ContextKeyDisplayName, name)
				}
				_ = mgr.TouchSession(c.Request().Context(), state.ID(), userID)
			}
			return next(c)
		}
	}
}

// RequireAuth protects routes by checking for a valid session user ID.
func RequireAuth(signinPath string) echo.MiddlewareFunc {
	return RequireAuthPath(func() string { return signinPath })
}

// RequireAuthPath protects routes by checking for a valid session user ID.
func RequireAuthPath(signinPath func() string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userID, _ := c.Get(ContextKeyUserID).(string)
			if userID == "" {
				path := signinPath()
				isHTMX := c.Request().Header.Get("HX-Request") == "true" &&
					c.Request().Header.Get("HX-Boosted") != "true"
				if isHTMX {
					c.Response().Header().Set("HX-Redirect", path)
					return c.NoContent(http.StatusUnauthorized)
				}
				return c.Redirect(http.StatusSeeOther, path)
			}
			return next(c)
		}
	}
}

// LoadUserRole resolves the authenticated user's role only for routes that
// actually need authorization data.
func LoadUserRole(users repository.UserReader) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userID, _ := c.Get(ContextKeyUserID).(string)
			if userID == "" {
				return next(c)
			}

			id, err := uuid.Parse(userID)
			if err != nil {
				return errUnauthorized("Your session is invalid. Please sign in again.")
			}

			user, err := users.GetByID(c.Request().Context(), id)
			if err != nil {
				if errors.Is(err, model.ErrUserNotFound) {
					return errUnauthorized("Your account could not be loaded. Please sign in again.")
				}
				return errUnavailable("Unable to authorize this request right now.", err)
			}

			c.Set(ContextKeyUserRole, user.Role)
			return next(c)
		}
	}
}

// RequireAdmin ensures the authenticated user has the admin role.
func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role, _ := c.Get(ContextKeyUserRole).(string)
			if role != model.RoleAdmin {
				return errForbidden("You are not allowed to access this area.")
			}
			return next(c)
		}
	}
}
