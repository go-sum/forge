package config

// Config is the root application configuration struct.
// App holds everything loaded from config/app.yaml.
// Site, Nav, and Service are loaded from their own optional files.
type Config struct {
	App     AppConfig     `koanf:"app"`
	Site    SiteConfig    `koanf:"site"`
	Nav     NavConfig     `koanf:"nav"`
	Service ServiceConfig `koanf:"service"`
}

// AppConfig holds the full application configuration from config/app.yaml.
type AppConfig struct {
	Env       string            `koanf:"env" validate:"required,oneof=development production test"`
	Name      string            `koanf:"name" validate:"required"`
	Server    ServerConfig      `koanf:"server"`
	Security  SecurityConfig    `koanf:"security"`
	Database  DatabaseConfig    `koanf:"database"`
	KV        KVConfig          `koanf:"kv"`
	Session   SessionConfig     `koanf:"session"`
	Auth      AuthConfig        `koanf:"auth"`
	Assets    AssetsConfig      `koanf:"assets"`
	Log       LogConfig         `koanf:"log"`
	CSPHashes CSPHashesConfig   `koanf:"csp_hashes"`
	Keys      ContextKeysConfig `koanf:"keys"`
}

type ServerConfig struct {
	Host            string `koanf:"host"             validate:"required"`
	Port            int    `koanf:"port"             validate:"required,min=1,max=65535"`
	GracefulTimeout int    `koanf:"graceful_timeout"`
}

type SecurityConfig struct {
	ExternalOrigin   string                     `koanf:"external_origin" validate:"required,url"`
	Origin           OriginConfig               `koanf:"origin"`
	FetchMetadata    FetchMetadataConfig        `koanf:"fetch_metadata"`
	Headers          HeadersConfig              `koanf:"headers"`
	CSRF             CSRFConfig                 `koanf:"csrf"`
	CORS             CORSConfig                 `koanf:"cors"`
	RateLimitBackend RateLimitBackendConfig     `koanf:"rate_limit_backend"`
	RateLimits       map[string]RateLimitConfig `koanf:"rate_limits"`
}

// CORSConfig configures the CORS middleware applied to opt-in route groups
// such as /api. All fields are optional — zero values mean CORS is permissive
// (wildcard origin, no credentials) which is safe for public API endpoints.
type CORSConfig struct {
	// AllowOrigins is the list of exact-match allowed origins.
	// When empty, the middleware uses "*" (wildcard); AllowCredentials must
	// be false when using wildcards.
	AllowOrigins []string `koanf:"allow_origins"`

	// AllowHeaders is the list of additional request headers permitted
	// cross-origin. Empty means headers are echoed from the preflight request.
	AllowHeaders []string `koanf:"allow_headers"`

	// AllowCredentials permits cookies and authorization headers cross-origin.
	// Cannot be combined with an empty AllowOrigins (wildcard).
	AllowCredentials bool `koanf:"allow_credentials"`

	// MaxAge is the preflight cache duration in seconds. 0 = header not sent.
	MaxAge int `koanf:"max_age"`
}

// RateLimitConfig configures the IP-based rate limiter applied to high-risk
// unauthenticated mutation routes (e.g. /signin, /signup).
// Rate 0 disables rate limiting entirely.
type RateLimitConfig struct {
	Rate      float64 `koanf:"rate"`       // requests per second (token bucket refill)
	Burst     int     `koanf:"burst"`      // maximum burst size above the steady rate
	ExpiresIn int     `koanf:"expires_in"` // seconds; 0 → store default (180s)
}

type RateLimitBackendConfig struct {
	Selected string `koanf:"selected" validate:"omitempty,oneof=memory"`
}

type OriginConfig struct {
	Enabled        bool     `koanf:"enabled"`
	RequireHeader  bool     `koanf:"require_header"`
	AllowedOrigins []string `koanf:"allowed_origins"`
}

type FetchMetadataConfig struct {
	Enabled                 bool     `koanf:"enabled"`
	AllowedSites            []string `koanf:"allowed_sites"`
	AllowedModes            []string `koanf:"allowed_modes"`
	AllowedDestinations     []string `koanf:"allowed_destinations"`
	FallbackWhenMissing     bool     `koanf:"fallback_when_missing"`
	RejectCrossSiteNavigate bool     `koanf:"reject_cross_site_navigate"`
}

type HeadersConfig struct {
	XSSProtection         string     `koanf:"xss_protection" validate:"required"`
	ContentTypeNosniff    bool       `koanf:"content_type_nosniff"`
	FrameOptions          string     `koanf:"frame_options" validate:"required"`
	ContentSecurityPolicy string     `koanf:"content_security_policy" validate:"required"`
	HSTS                  HSTSConfig `koanf:"hsts"`
}

type HSTSConfig struct {
	Enabled           bool `koanf:"enabled"`
	MaxAge            int  `koanf:"max_age"`
	IncludeSubDomains bool `koanf:"include_subdomains"`
	Preload           bool `koanf:"preload"`
}

type CSRFConfig struct {
	Key        string `koanf:"key"         validate:"required,min=32"`
	FormField  string `koanf:"form_field"  validate:"required"`
	HeaderName string `koanf:"header_name" validate:"required"`
	TokenTTL   int    `koanf:"token_ttl"   validate:"omitempty,min=1"` // seconds; defaults to 3600 when unset
}

type DatabaseConfig struct {
	URL string `koanf:"url" validate:"required"`
}

// KVConfig holds the key-value store configuration.
type KVConfig struct {
	Enabled bool         `koanf:"enabled"`
	Store   string       `koanf:"store" validate:"omitempty,oneof=redis"`
	Redis   RedisKVConfig `koanf:"redis"`
}

// RedisKVConfig holds Redis/Dragonfly connection parameters.
type RedisKVConfig struct {
	Addr         string `koanf:"addr"           validate:"required_if=Enabled true"`
	Password     string `koanf:"password"`
	DB           int    `koanf:"db"`
	PoolSize     int    `koanf:"pool_size"`
	MinIdleConns int    `koanf:"min_idle_conns"`
	DialTimeout  int    `koanf:"dial_timeout"`  // seconds
	ReadTimeout  int    `koanf:"read_timeout"`  // seconds
	WriteTimeout int    `koanf:"write_timeout"` // seconds
}

type AssetsConfig struct {
	PublicDir    string `koanf:"public_dir"`
	PublicPrefix string `koanf:"public_prefix"`
}

type AuthConfig struct{}

type SessionConfig struct {
	Store      string `koanf:"store"       validate:"omitempty,oneof=cookie server"`
	Name       string `koanf:"name"`
	AuthKey    string `koanf:"auth_key"    validate:"required,min=32"`
	EncryptKey string `koanf:"encrypt_key" validate:"required,min=16,max=32"`
	MaxAge     int    `koanf:"max_age"`
	Secure     bool   `koanf:"secure"`
}

type LogConfig struct {
	Level string `koanf:"level" validate:"required,oneof=debug info warn error"`
}

type CSPHashesConfig struct {
	Always  []string `koanf:"always"`
	DevOnly []string `koanf:"dev_only"`
}

// ContextKeysConfig defines the Echo context key names written by auth middleware
// and read by the view layer. Override in config/app.yaml under app.keys.
type ContextKeysConfig struct {
	UserID      string `koanf:"user_id"`
	UserRole    string `koanf:"user_role"`
	DisplayName string `koanf:"display_name"`
	CSRF        string `koanf:"csrf"`
}
