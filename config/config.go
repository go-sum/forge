// Defines the application configuration schema and loader.
package config

import (
	"github.com/go-playground/validator/v10"
	uilayout "github.com/go-sum/componentry/ui/layout"
	auth "github.com/go-sum/auth"
	cfgs "github.com/go-sum/server/config"
)

// Config is the root application configuration struct.
type Config struct {
	App      AppConfig
	Nav      NavConfig
	Security SecurityConfig
	Service  ServiceConfig
	Session  SessionsConfig
	Site     SiteConfig
	Store    StoreConfig
}

// App is the global configuration singleton, initialised at startup.
var App *Config

// environmentMap maps environment names to their ordered overlay functions.
var environmentMap = map[string][]func(*Config){
	"development": {developmentConfig},
}

// Load loads the application configuration.
func Load(appEnv string) (*Config, error) {
	return cfgs.Load(productionDefault, override(appEnv)...)
}

// productionDefault returns a fully populated Config with production defaults.
// All values are Go literals; secrets use ExpandEnv for env var injection.
func productionDefault() Config {
	return Config{
		App:      defaultApp(),
		Nav:      defaultNav(),
		Security: defaultSecurity(),
		Service:  defaultService(),
		Session:  defaultSession(),
		Site:     defaultSite(),
		Store:    defaultStore(),
	}
}

// developmentConfig applies development-mode configuration.
func developmentConfig(cfg *Config) {
	cfg.App.Env = "development"
	cfg.App.Log.Level = "debug"
	cfg.App.Server.Port = 3000
	cfg.Session.Auth.Secure = false
	cfg.Session.Auth.Store = "server"
	cfg.Security.CSPHashes.DevOnly = []string{"'sha256-y933zYNvpVe5f9j5A+OKECUXiWo8bKB5Yp5sLDD3d0I='"}
	cfg.Security.ExternalOrigin = "https://forge.test"
	cfg.Security.Headers.ContentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; object-src 'none'; base-uri 'self'"
	cfg.Security.Headers.HSTS.Enabled = false
	cfg.Security.RateLimits["auth"] = RateLimitConfig{Rate: 0.2, Burst: 5, ExpiresIn: 300}
	cfg.Store.Database.AutoMigrate = true
	cfg.Store.KV.Enabled = true
	cfg.Service.Auth.Methods.Passkey = auth.PasskeyMethodConfig{
		Enabled:       true,
		RPDisplayName: "Forge (dev)",
		RPID:          "forge.test",
		RPOrigins:     []string{"https://forge.test"},
	}
}

// Returns the ordered overlay functions for the given environment.
func override(appEnv string) []func(*Config) {
	return environmentMap[appEnv]
}

// RegisterValidationRules composes all cross-field validation rule-sets that apply to this configuration.
func (c *Config) RegisterValidationRules(v *validator.Validate) {
	v.RegisterStructValidation(crossFieldRules, Config{})
	c.Service.Auth.RegisterValidationRules(v)
	uilayout.RegisterNavValidations(v)
}

// crossFieldRules validates cross-field invariants on Config.
func crossFieldRules(sl validator.StructLevel) {
	cfg := sl.Current().Interface().(Config)

	// Server-side session store requires the KV store to be enabled.
	if cfg.Session.Auth.Store == "server" && !cfg.Store.KV.Enabled {
		sl.ReportError(cfg.Session.Auth.Store, "Session.Auth.Store", "Store", "requires_kv", "")
	}
}
