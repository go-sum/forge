// Package ctxkeys defines typed context keys for values stored in request contexts.
// Using a private type prevents external packages from colliding with these keys.
package ctxkeys

type ctxKey string

const (
	UserID          ctxKey = "user_id"
	UserEmail       ctxKey = "user_email"
	UserRole        ctxKey = "user_role"
	IsAuthenticated ctxKey = "is_authenticated"
	RequestID       ctxKey = "request_id"
	Logger          ctxKey = "logger"
	CSRF            ctxKey = "csrf"
	Config          ctxKey = "config"
)
