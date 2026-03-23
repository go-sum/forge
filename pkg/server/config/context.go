package config

import "github.com/labstack/echo/v5"

// Set stores val in c under key.
func Set(c *echo.Context, key string, val any) {
	c.Set(key, val)
}

// Get retrieves the value stored under key, type-asserting to T.
// Returns the zero value and false if the key is absent or the type does not match.
func Get[T any](c *echo.Context, key string) (T, bool) {
	v, ok := c.Get(key).(T)
	return v, ok
}
