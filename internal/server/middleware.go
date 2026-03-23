// Package server wires the application's middleware stack onto an Echo instance.
// This is the application-level configuration point — edit here to add, remove,
// or reorder middleware for this specific application.
//
// The separation from pkg/server is intentional:
//   - pkg/server.New() creates a bare Echo instance (generic, extractable)
//   - internal/server.RegisterMiddleware() configures this app's specific middleware (edit freely)
package server

import (
	"log/slog"
	"net/http"

	"github.com/go-sum/forge/config"
	srv "github.com/go-sum/server"
	smw "github.com/go-sum/server/middleware"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// RegisterMiddleware wires the application middleware stack onto e in the correct order.
// cfg carries the runtime values (CSP policy, CSRF cookie name, public prefix,
// cookie security flag) that middleware need to be configured with.
// appCfg is forwarded to the error handler so error pages render the correct nav
// and CSRF token is stored under the same key that the view layer reads.
func RegisterMiddleware(e *echo.Echo, cfg srv.Config, appCfg *config.Config) {
	e.HTTPErrorHandler = NewErrorHandler(ErrorHandlerConfig{
		Debug:     cfg.Debug,
		Logger:    slog.Default(),
		NavConfig: appCfg.Nav,
		Keys:      appCfg.Keys,
	})

	// Pre-routing: runs before the router dispatches the request.
	e.Pre(middleware.RemoveTrailingSlash())

	// Post-routing (order matters — each middleware wraps the next).
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// HSTS is only meaningful over TLS. Only emit it when CookieSecure is true
	// (which correlates with HTTPS / production), so dev logs stay clean.
	const hstsOneYear = 31536000
	hstsMaxAge := 0
	if cfg.CookieSecure {
		hstsMaxAge = hstsOneYear
	}
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		// X-XSS-Protection: 1; mode=block is obsolete and actively harmful in
		// some legacy browsers. OWASP recommends "0"; CSP handles real protection.
		XSSProtection:         "0",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            hstsMaxAge,
		HSTSPreloadEnabled:    cfg.CookieSecure,
		ContentSecurityPolicy: cfg.CSP,
	}))

	// Log 5xx as Error, 4xx as Warn, 2xx/3xx only in debug mode.
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []any{
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency,
				"remote_ip", v.RemoteIP,
				"request_id", v.RequestID,
			}
			ctx := c.Request().Context()
			switch {
			case v.Status >= 500:
				slog.ErrorContext(ctx, "REQUEST", attrs...)
			case v.Status >= 400:
				slog.WarnContext(ctx, "REQUEST", attrs...)
			default:
				if !slog.Default().Enabled(ctx, slog.LevelDebug) {
					return nil // suppress 2xx/3xx in non-debug mode
				}
				slog.InfoContext(ctx, "REQUEST", attrs...)
			}
			return nil
		},
	}))

	// CSRF runs after the logger so all requests (including rejections) are logged.
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		ContextKey:     appCfg.Keys.CSRF,
		TokenLookup:    "header:X-CSRF-Token,form:" + cfg.CSRFCookieName,
		CookieName:     cfg.CSRFCookieName,
		CookieSameSite: http.SameSiteLaxMode,
		CookieSecure:   cfg.CookieSecure,
		CookiePath:     "/",
	}))

	e.Use(smw.StaticCacheControl(cfg.PublicPrefix))
}
