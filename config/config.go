// Package config defines the application's configuration schema.
// Types are split by their YAML source: config_types.go, site_types.go, nav_types.go.
// Configuration is loaded at startup by internal/app.
package config

import (
	"strings"

	"github.com/go-playground/validator/v10"

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
				{Filepath: dir + "/config.yaml", Target: &cfg.App},
				{Filepath: dir + "/site.yaml", Target: &cfg.Site},
				{Filepath: dir + "/nav.yaml", Target: &cfg.Nav, Validator: RegisterNavValidations},
				{Filepath: dir + "/service.yaml", Target: &cfg.Service},
			},
		}
	})
}

// RegisterNavValidations registers the declarative nav schema rules on v.
func RegisterNavValidations(v *validator.Validate) {
	v.RegisterStructValidation(navItemStructValidation, NavItem{})
}

func navItemStructValidation(sl validator.StructLevel) {
	item := sl.Current().Interface().(NavItem)

	if item.Type == "separator" {
		reportIfSet(sl, item.Slot, "Slot", "slot", "separator_only")
		reportIfSet(sl, item.Label, "Label", "label", "separator_only")
		reportIfSet(sl, item.Href, "Href", "href", "separator_only")
		reportIfSet(sl, item.Action, "Action", "action", "separator_only")
		reportIfSet(sl, item.Method, "Method", "method", "separator_only")
		reportIfSet(sl, item.Icon, "Icon", "icon", "separator_only")
		reportIfTrue(sl, item.MatchPrefix, "MatchPrefix", "match_prefix", "separator_only")
		reportIfLen(sl, len(item.HiddenFields), "HiddenFields", "hidden_fields", "separator_only")
		reportIfLen(sl, len(item.Items), "Items", "items", "separator_only")
		return
	}

	if item.MatchPrefix && item.Href == "" {
		sl.ReportError(item.MatchPrefix, "MatchPrefix", "match_prefix", "requires_href", "")
	}
	if item.Method != "" && item.Action == "" {
		sl.ReportError(item.Method, "Method", "method", "requires_action", "")
	}
	if len(item.HiddenFields) > 0 && item.Action == "" {
		sl.ReportError(item.HiddenFields, "HiddenFields", "hidden_fields", "requires_action", "")
	}

	if item.Slot != "" {
		reportIfSet(sl, item.Href, "Href", "href", "slot_conflict")
		reportIfSet(sl, item.Action, "Action", "action", "slot_conflict")
		reportIfSet(sl, item.Method, "Method", "method", "slot_conflict")
		reportIfTrue(sl, item.MatchPrefix, "MatchPrefix", "match_prefix", "slot_conflict")
		reportIfLen(sl, len(item.HiddenFields), "HiddenFields", "hidden_fields", "slot_conflict")
		reportIfLen(sl, len(item.Items), "Items", "items", "slot_conflict")
		return
	}

	hasHref := item.Href != ""
	hasAction := item.Action != ""
	hasItems := len(item.Items) > 0

	if hasHref && hasAction {
		sl.ReportError(item.Action, "Action", "action", "conflicts_with_href", "")
	}
	if hasAction && hasItems {
		sl.ReportError(item.Action, "Action", "action", "conflicts_with_items", "")
	}

	if (hasHref || hasAction || hasItems) && item.Label == "" {
		sl.ReportError(item.Label, "Label", "label", "required_for_item", "")
	}

	if !hasHref && !hasAction && !hasItems && item.Label == "" {
		sl.ReportError(item.Label, "Label", "label", "required", "")
	}
}

func reportIfSet(sl validator.StructLevel, value string, fieldName, jsonName, tag string) {
	if value != "" {
		sl.ReportError(value, fieldName, jsonName, tag, "")
	}
}

func reportIfTrue(sl validator.StructLevel, value bool, fieldName, jsonName, tag string) {
	if value {
		sl.ReportError(value, fieldName, jsonName, tag, "")
	}
}

func reportIfLen(sl validator.StructLevel, n int, fieldName, jsonName, tag string) {
	if n > 0 {
		sl.ReportError(n, fieldName, jsonName, tag, "")
	}
}
