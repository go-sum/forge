// Package config is the application configuration package.
// Compile-time values (EnvPrefix, App singleton) live here.
// Type definitions live in types.go. Nav types and validation live in nav.go.
// Runtime values are loaded from config.yaml, environment overlays, and CTX_-prefixed
// environment variables
package config

import cfgs "github.com/go-sum/server/config"

// EnvPrefix is the compile-time environment variable prefix for this application.
// Environment variables with this prefix are mapped to config keys after stripping
// the prefix and lowercasing (e.g. CTX_SERVER_PORT → server.port).
const EnvPrefix = "CTX_"

// App is the global configuration singleton, populated by InitConfig.
var App *Config

// InitConfig loads configuration from YAML files and CTX_-prefixed env vars in baseDir
// into the App singleton. Panicking on misconfiguration is the caller's responsibility
// (see internal/app/infra.go).
func InitConfig(baseDir string) error {
	cfg, err := cfgs.Load(func(cfg *Config) cfgs.Options {
		return cfgs.Options{
			EnvPrefix:      EnvPrefix,
			BaseDir:        baseDir,
			EnvKey:         "app.env",
			ValidatorSetup: RegisterNavValidations,
			ContentFiles: []cfgs.ContentFile{
				{Filename: "site.yaml", Target: &cfg.Site},
				{Filename: "nav.yaml", Target: &cfg.Nav},
			},
		}
	})
	if err != nil {
		return err
	}
	App = cfg
	return nil
}
