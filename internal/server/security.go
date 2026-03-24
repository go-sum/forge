package server

import (
	"github.com/go-sum/forge/config"
	"github.com/go-sum/security/fetchmeta"
	"github.com/go-sum/security/headers"
	secmw "github.com/go-sum/security/middleware"
	"github.com/go-sum/security/origin"
	"github.com/labstack/echo/v5"
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

func securityHeaderPolicy(cfg *config.Config, processedCSP string) headers.Policy {
	h := cfg.Security.Headers
	return headers.Policy{
		XSSProtection:         h.XSSProtection,
		ContentTypeNosniff:    h.ContentTypeNosniff,
		FrameOptions:          h.FrameOptions,
		ContentSecurityPolicy: processedCSP,
		HSTS: headers.HSTSConfig{
			Enabled:           h.HSTS.Enabled,
			MaxAge:            h.HSTS.MaxAge,
			IncludeSubDomains: h.HSTS.IncludeSubDomains,
			Preload:           h.HSTS.Preload,
		},
	}
}
