// Package server wires the application's middleware stack onto an Echo instance.
// This is the application-level configuration point — edit here to add, remove,
// or reorder middleware for this specific application.
//
// The separation from pkg/server is intentional:
//   - pkg/server.NewWithConfig() creates an Echo instance with construction-time hooks (generic, extractable)
//   - internal/server.RegisterMiddleware() configures this app's specific middleware (edit freely)
package server

import (
	"log/slog"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/server/middleware/override"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// RegisterMiddleware wires the application middleware stack onto e in the correct order.
// The HTTPErrorHandler must be set at construction time via server.NewWithConfig.
// processedCSP is the final CSP header value (with hashes already injected by the caller).
func RegisterMiddleware(e *echo.Echo, cfg *config.Config, processedCSP string) {
	// Pre-routing: runs before the router dispatches the request.
	e.Pre(middleware.RemoveTrailingSlash())
	// Method override reads _method from POST bodies and promotes the request
	// method to PUT, PATCH, or DELETE before routing. Must run in Pre so the
	// router sees the promoted method. Any disallowed override value returns 400.
	e.Pre(override.Middleware())

	// Post-routing (order matters — each middleware wraps the next).
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(secureMiddleware(cfg, processedCSP))
	e.Use(secureHeaders(cfg))

	// Log 5xx as Error, 4xx as Warn, 2xx/3xx only in debug mode.
	// Each Log* flag must be explicitly enabled — Echo v5 opts out of capture by default.
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		// HandleError:  true,
		// LogMethod:    true,
		// LogURI:       true,
		// LogStatus:    true,
		// LogLatency:   true,
		// LogRemoteIP:  true,
		// LogRequestID: true,
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
	e.Use(CSRFMiddleware(cfg))
}
