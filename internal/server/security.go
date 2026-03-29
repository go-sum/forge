package server

import (
	"github.com/go-sum/forge/config"
	"github.com/go-sum/security/cors"
	"github.com/go-sum/security/csrf"
	"github.com/go-sum/security/fetchmeta"
	secmw "github.com/go-sum/security/middleware"
	"github.com/go-sum/security/origin"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// CrossOriginGuard applies origin and Fetch Metadata checks to unsafe requests.
func CrossOriginGuard(cfg *config.Config) echo.MiddlewareFunc {
	originPolicy := origin.Policy{
		Enabled:         cfg.App.Security.Origin.Enabled,
		CanonicalOrigin: cfg.App.Security.ExternalOrigin,
		RequireHeader:   cfg.App.Security.Origin.RequireHeader,
		AllowedOrigins:  cfg.App.Security.Origin.AllowedOrigins,
	}
	fetchPolicy := fetchmeta.Policy{
		Enabled:                 cfg.App.Security.FetchMetadata.Enabled,
		AllowedSites:            cfg.App.Security.FetchMetadata.AllowedSites,
		AllowedModes:            cfg.App.Security.FetchMetadata.AllowedModes,
		AllowedDestinations:     cfg.App.Security.FetchMetadata.AllowedDestinations,
		FallbackWhenMissing:     cfg.App.Security.FetchMetadata.FallbackWhenMissing,
		RejectCrossSiteNavigate: cfg.App.Security.FetchMetadata.RejectCrossSiteNavigate,
	}

	return secmw.CrossOriginGuard(originPolicy, fetchPolicy)
}

// CSRFMiddleware applies HMAC-signed, time-limited CSRF protection with typed
// errors so that token failures are rendered as 403 Forbidden by the app's
// error handler rather than falling through to a 500 Internal Server Error.
func CSRFMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	c := cfg.App.Security.CSRF
	return csrf.Middleware(csrf.Config{
		Key:        []byte(c.Key),
		TokenTTL:   c.TokenTTL,
		ContextKey: cfg.App.Keys.CSRF,
		HeaderName: c.HeaderName,
		FormField:  c.FormField,
	})
}

// CORSMiddleware returns a CORS middleware for opt-in route groups (e.g. /api).
// When AllowOrigins is empty the middleware permits all origins via "*";
// AllowCredentials must be false in that case, which is the default.
func CORSMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	c := cfg.App.Security.CORS
	allowOrigins := c.AllowOrigins
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"*"}
	}
	return cors.Middleware(cors.Config{
		Mode:             cors.OriginModeExact,
		AllowOrigins:     allowOrigins,
		AllowHeaders:     c.AllowHeaders,
		AllowCredentials: c.AllowCredentials,
		MaxAge:           c.MaxAge,
	})
}

// secureMiddleware applies HTTP security headers via Echo's built-in Secure
// middleware. HSTS is TLS-conditional (not written on plain HTTP in dev).
func secureMiddleware(cfg *config.Config, processedCSP string) echo.MiddlewareFunc {
	h := cfg.App.Security.Headers
	nosniff := ""
	if h.ContentTypeNosniff {
		nosniff = "nosniff"
	}
	hstsMaxAge := 0
	if h.HSTS.Enabled {
		hstsMaxAge = h.HSTS.MaxAge
	}
	return middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         h.XSSProtection,
		ContentTypeNosniff:    nosniff,
		XFrameOptions:         h.FrameOptions,
		ContentSecurityPolicy: processedCSP,
		HSTSMaxAge:            hstsMaxAge,
		HSTSExcludeSubdomains: !h.HSTS.IncludeSubDomains,
		HSTSPreloadEnabled:    h.HSTS.Preload,
	})
}
