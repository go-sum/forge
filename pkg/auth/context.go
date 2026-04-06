package auth

import "github.com/labstack/echo/v5"

// Context key constants for auth state stored in the Echo context.
// These replace the former configurable ContextKeys struct.
const (
	ContextKeyUserID      = "auth.user_id"
	ContextKeyUserRole    = "auth.user_role"
	ContextKeyDisplayName = "auth.display_name"
)

// UserID extracts the authenticated user's ID from the Echo context.
func UserID(c *echo.Context) string {
	v, _ := c.Get(ContextKeyUserID).(string)
	return v
}

// UserRole extracts the authenticated user's role from the Echo context.
func UserRole(c *echo.Context) string {
	v, _ := c.Get(ContextKeyUserRole).(string)
	return v
}

// DisplayName extracts the authenticated user's display name from the Echo context.
func DisplayName(c *echo.Context) string {
	v, _ := c.Get(ContextKeyDisplayName).(string)
	return v
}
