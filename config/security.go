package config

import cfgs "github.com/go-sum/server/config"

type SecurityConfig struct {
	ExternalOrigin   string `validate:"required,url"`
	Origin           OriginConfig
	FetchMetadata    FetchMetadataConfig
	Headers          HeadersConfig
	CSRF             CSRFConfig
	CORS             CORSConfig
	RateLimitBackend RateLimitBackendConfig
	RateLimits       map[string]RateLimitConfig
	CSPHashes        CSPHashesConfig
}

type CSPHashesConfig struct {
	Always  []string
	DevOnly []string
}

// CORSConfig configures the CORS middleware applied to opt-in route groups
// such as /api. All fields are optional — zero values mean CORS is permissive
// (wildcard origin, no credentials) which is safe for public API endpoints.
type CORSConfig struct {
	// AllowOrigins is the list of exact-match allowed origins.
	// When empty, the middleware uses "*" (wildcard); AllowCredentials must
	// be false when using wildcards.
	AllowOrigins []string

	// AllowHeaders is the list of additional request headers permitted
	// cross-origin. Empty means headers are echoed from the preflight request.
	AllowHeaders []string

	// AllowCredentials permits cookies and authorization headers cross-origin.
	// Cannot be combined with an empty AllowOrigins (wildcard).
	AllowCredentials bool

	// MaxAge is the preflight cache duration in seconds. 0 = header not sent.
	MaxAge int
}

// RateLimitConfig configures the IP-based rate limiter applied to high-risk
// unauthenticated mutation routes (e.g. /signin, /signup).
// Rate 0 disables rate limiting entirely.
type RateLimitConfig struct {
	Rate      float64 // requests per second (token bucket refill)
	Burst     int     // maximum burst size above the steady rate
	ExpiresIn int     // seconds; 0 → store default (180s)
}

type RateLimitBackendConfig struct {
	Selected string `validate:"required,oneof=memory"`
}

type OriginConfig struct {
	Enabled        bool
	RequireHeader  bool
	AllowedOrigins []string
}

type FetchMetadataConfig struct {
	Enabled                 bool
	AllowedSites            []string
	AllowedModes            []string
	AllowedDestinations     []string
	FallbackWhenMissing     bool
	RejectCrossSiteNavigate bool
}

type HeadersConfig struct {
	XSSProtection         string `validate:"required"`
	ContentTypeNosniff    bool
	FrameOptions          string `validate:"required"`
	ContentSecurityPolicy string `validate:"required"`
	ReferrerPolicy        string
	PermissionsPolicy     string
	HSTS                  HSTSConfig
}

type HSTSConfig struct {
	Enabled           bool
	MaxAge            int
	IncludeSubDomains bool
	Preload           bool
}

type CSRFConfig struct {
	Key        string `validate:"required,min=32"`
	ContextKey string `validate:"required"`
	FormField  string `validate:"required"`
	HeaderName string `validate:"required"`
	TokenTTL   int    `validate:"omitempty,min=1"` // seconds; defaults to 3600 when unset
}

func defaultSecurity() SecurityConfig {
	return SecurityConfig{
		ExternalOrigin: cfgs.ExpandEnv("${EXTERNAL_ORIGIN}"),
		Origin: OriginConfig{
			Enabled:        true,
			RequireHeader:  true,
			AllowedOrigins: []string{},
		},
		FetchMetadata: FetchMetadataConfig{
			Enabled:                 true,
			AllowedSites:            []string{"same-origin", "same-site"},
			AllowedModes:            []string{"cors", "navigate", "same-origin"},
			AllowedDestinations:     []string{},
			FallbackWhenMissing:     true,
			RejectCrossSiteNavigate: true,
		},
		Headers: HeadersConfig{
			XSSProtection:         "0",
			ContentTypeNosniff:    true,
			FrameOptions:          "DENY",
			ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self'; font-src 'self'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; object-src 'none'; base-uri 'self'",
			ReferrerPolicy:        "strict-origin-when-cross-origin",
			PermissionsPolicy:     "camera=(), microphone=(), geolocation=()",
			HSTS: HSTSConfig{
				Enabled:           true,
				MaxAge:            31536000,
				IncludeSubDomains: true,
				Preload:           true,
			},
		},
		CSRF: CSRFConfig{
			Key:        cfgs.ExpandEnv("${SECURITY_CSRF_KEY}"),
			ContextKey: "csrf",
			FormField:  "_csrf",
			HeaderName: "X-CSRF-Token",
			TokenTTL:   3600,
		},
		CORS: CORSConfig{
			AllowOrigins:     []string{},
			AllowHeaders:     []string{},
			AllowCredentials: false,
			MaxAge:           0,
		},
		RateLimitBackend: RateLimitBackendConfig{
			Selected: "memory",
		},
		RateLimits: map[string]RateLimitConfig{
			"server": {Rate: 50, Burst: 100, ExpiresIn: 180},
			"auth":   {Rate: 2, Burst: 5, ExpiresIn: 300},
		},
		CSPHashes: CSPHashesConfig{
			Always:  []string{},
			DevOnly: []string{},
		},
	}
}
