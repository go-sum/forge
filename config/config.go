// Package config defines the application's configuration schema.
// Type definitions live in types.go. Nav types and validation live in nav.go.
// Configuration is loaded at startup by internal/app.
package config

// EnvPrefix is the environment variable prefix for this application.
// Variables with this prefix are mapped to config keys after stripping the prefix
// and lowercasing (e.g. CTX_SERVER_PORT → server.port).
const EnvPrefix = "CTX_"

// App is the global configuration singleton, initialised at startup.
var App *Config
