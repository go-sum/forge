// Package config defines the application's configuration schema.
// Types are split by their YAML source: app.go, site.go, nav.go, service.go.
// Configuration is loaded at startup by internal/app.
package config

import (
	"strings"

	cfgs "github.com/go-sum/server/config"
)

// App is the global configuration singleton, initialised at startup.
var App *Config

// Environment returns c.App.Env lowercased, defaulting to "production" when empty.
func (c *Config) Environment() string {
	if c.App.Env != "" {
		return strings.ToLower(c.App.Env)
	}
	return "production"
}

// IsDevelopment reports whether the application is running in development mode.
func (c *Config) IsDevelopment() bool { return c.Environment() == "development" }

// IsProduction reports whether the application is running in production mode.
func (c *Config) IsProduction() bool { return c.Environment() == "production" }

// DSN is an alias for App.Database.URL.
func (c *Config) DSN() string { return c.App.Database.URL }

// Load loads the application configuration from the default config/ directory.
// appEnv is typically os.Getenv("APP_ENV").
func Load(appEnv string) (*Config, error) {
	return LoadFrom("config", appEnv)
}

// LoadFrom loads configuration from the given directory.
// It is the primary entry point for both production and test use.
func LoadFrom(dir, appEnv string) (*Config, error) {
	return cfgs.Load(func(cfg *Config) cfgs.Options {
		return cfgs.Options{
			EnvKey: appEnv,
			Files: []cfgs.ConfigFile{
				{Filepath: dir + "/app.yaml"},
				{Filepath: dir + "/site.yaml"},
				{Filepath: dir + "/nav.yaml"},
				{Filepath: dir + "/service.yaml"},
			},
		}
	})
}
