package server

import (
	"github.com/go-sum/forge/config"
	"github.com/go-sum/security/csrf"
	"github.com/go-sum/security/fetchmeta"
	secmw "github.com/go-sum/security/middleware"
	"github.com/go-sum/security/origin"
	"github.com/go-sum/security/ratelimit"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// ProtectBrowserMutation applies origin and Fetch Metadata checks to unsafe requests.
func ProtectBrowserMutation(cfg *config.Config) echo.MiddlewareFunc {
	originPolicy := origin.Policy{
		Enabled:         cfg.Security.Origin.Enabled,
		CanonicalOrigin: cfg.Security.ExternalOrigin,
		RequireHeader:   cfg.Security.Origin.RequireHeader,
		AllowedOrigins:  cfg.Security.Origin.AllowedOrigins,
	}
	fetchPolicy := fetchmeta.Policy{
		Enabled:                 cfg.Security.FetchMetadata.Enabled,
		AllowedSites:            cfg.Security.FetchMetadata.AllowedSites,
		AllowedModes:            cfg.Security.FetchMetadata.AllowedModes,
		AllowedDestinations:     cfg.Security.FetchMetadata.AllowedDestinations,
		FallbackWhenMissing:     cfg.Security.FetchMetadata.FallbackWhenMissing,
		RejectCrossSiteNavigate: cfg.Security.FetchMetadata.RejectCrossSiteNavigate,
	}

	return secmw.ProtectBrowserMutation(originPolicy, fetchPolicy)
}

// CSRFMiddleware applies Echo's double-submit cookie CSRF protection with typed
// errors so that token failures are rendered as 403 Forbidden by the app's
// error handler rather than falling through to a 500 Internal Server Error.
func CSRFMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	c := cfg.Security.CSRF
	return csrf.Middleware(csrf.Config{
		ContextKey:   cfg.Keys.CSRF,
		HeaderName:   c.HeaderName,
		FormField:    c.FormField,
		CookieName:   c.CookieName,
		CookieSecure: cfg.Auth.Session.Secure,
	})
}

// secureMiddleware applies HTTP security headers via Echo's built-in Secure
// middleware. HSTS is TLS-conditional (not written on plain HTTP in dev).
func secureMiddleware(cfg *config.Config, processedCSP string) echo.MiddlewareFunc {
	h := cfg.Security.Headers
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

// RateLimitMiddleware applies IP-based rate limiting for the named policy from
// cfg.Security.RateLimits. When the policy is absent or Rate is 0, it returns a
// no-op passthrough so it can be composed inline without nil guards:
//
//	e.Group("", appserver.RateLimitMiddleware(c.Config, "server"))
//	publicMutations.Use(appserver.ProtectBrowserMutation(c.Config), appserver.RateLimitMiddleware(c.Config, "auth"))
//
// Each call creates an independent in-memory store, so different policy names
// maintain completely separate per-IP bucket maps.
func RateLimitMiddleware(cfg *config.Config, name string) echo.MiddlewareFunc {
	rl, ok := cfg.Security.RateLimits[name]
	if !ok || rl.Rate == 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}
	return ratelimit.Middleware(ratelimit.Config{
		Rate:  rl.Rate,
		Burst: rl.Burst,
	})
}
