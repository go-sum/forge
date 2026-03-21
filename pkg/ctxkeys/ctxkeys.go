// Package ctxkeys defines typed context keys for values stored in request contexts.
// Using a private type prevents external packages from colliding with these keys.
package ctxkeys

type ctxKey string

const (
	UserID          ctxKey = "user_id"
	UserRole        ctxKey = "user_role"
	UserDisplayName ctxKey = "user_display_name"
)
