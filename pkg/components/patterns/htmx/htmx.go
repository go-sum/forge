// Package htmx provides helpers for reading HTMX request headers and writing
// HTMX response headers in Echo v5 handlers.
package htmx

import "github.com/labstack/echo/v5"

// Request inspection helpers.

func IsRequest(c *echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

func IsBoosted(c *echo.Context) bool {
	return c.Request().Header.Get("HX-Boosted") == "true"
}

func GetTrigger(c *echo.Context) string {
	return c.Request().Header.Get("HX-Trigger")
}

func GetTarget(c *echo.Context) string {
	return c.Request().Header.Get("HX-Target")
}

func GetTriggerName(c *echo.Context) string {
	return c.Request().Header.Get("HX-Trigger-Name")
}

func GetCurrentURL(c *echo.Context) string {
	return c.Request().Header.Get("HX-Current-URL")
}

// Response header helpers.

func SetRedirect(c *echo.Context, url string) {
	c.Response().Header().Set("HX-Redirect", url)
}

func SetRefresh(c *echo.Context) {
	c.Response().Header().Set("HX-Refresh", "true")
}

func SetPushURL(c *echo.Context, url string) {
	c.Response().Header().Set("HX-Push-Url", url)
}

func SetReplaceURL(c *echo.Context, url string) {
	c.Response().Header().Set("HX-Replace-Url", url)
}

func SetTrigger(c *echo.Context, event string) {
	c.Response().Header().Set("HX-Trigger", event)
}

func SetTriggerAfterSettle(c *echo.Context, event string) {
	c.Response().Header().Set("HX-Trigger-After-Settle", event)
}

func SetRetarget(c *echo.Context, selector string) {
	c.Response().Header().Set("HX-Retarget", selector)
}

func SetReswap(c *echo.Context, strategy string) {
	c.Response().Header().Set("HX-Reswap", strategy)
}
