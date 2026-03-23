// Package config provides a generic, reusable configuration loader built on
// koanf. It supports layered loading (YAML files + environment variables),
// smart env-key transformation that handles underscores in field names, and
// struct validation via go-playground/validator tags.
//
// This package is a leaf node: it imports only external modules and the
// standard library — never anything from internal/ or other pkg/ packages.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	koanyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// ContentFile is a config overlay file with an optional validation scope.
// Validation errors from Target are attributed to Filename for precise diagnostics.
type ContentFile struct {
	// Filename is relative to BaseDir; missing files are silently skipped.
	Filename string
	// Target is a pointer into the already-unmarshalled root struct (e.g. &cfg.Site).
	// When non-nil, v.Struct(Target) runs after unmarshal and errors name this file.
	Target any
}

// Options configures how Load discovers and merges configuration sources.
type Options struct {
	// EnvPrefix is stripped from env var names before key mapping.
	// "CTX_" maps CTX_SERVER_PORT → server.port.
	EnvPrefix string

	BaseDir string

	// EnvKey selects the per-environment overlay file.
	// "app.env" causes Load to read "config.development.yaml" when app.env = "development".
	EnvKey string

	// ContentFiles are loaded after env vars, so their values intentionally win
	// over environment variables (site titles, logo paths, etc. belong to the app,
	// not the deployment). Missing files are silently skipped.
	ContentFiles []ContentFile

	// ValidatorSetup registers any custom validations needed by the caller's schema.
	ValidatorSetup func(v *validator.Validate)
}

// loadConfig merges configuration from multiple sources into target and validates it.
//
// Loading order (last writer wins, except ContentFiles which load after env vars):
//  1. BaseDir/config.yaml               — required; error if missing
//  2. BaseDir/config.{env}.yaml         — optional env overlay; silently skipped
//  3. EnvPrefix-prefixed env vars       — highest precedence for operational config
//  4. BaseDir/ContentFiles[*].Filename  — optional; intentionally wins over env vars
//     for content (title, logo, etc. belong to the app, not deployment)
//  5. Unmarshal into target
//  6. Per-file validation: for each ContentFile with non-nil Target, v.Struct(Target)
//     with the filename named in any error
//  7. Validate target using go-playground/validator struct tags
func loadConfig(target any, opts Options) error {
	k := koanf.New(".")

	// 1. Base config (required).
	baseFile := filepath.Join(opts.BaseDir, "config.yaml")
	if err := k.Load(file.Provider(baseFile), koanyaml.Parser()); err != nil {
		return fmt.Errorf("config: load %s: %w", baseFile, err)
	}

	// 2. Environment overlay (optional — silently skip if file is absent).
	// Peek at the corresponding env var first (e.g. CTX_APP_ENV) so it can
	// drive overlay selection even when config.yaml sets a different default.
	if opts.EnvKey != "" {
		envVarName := opts.EnvPrefix + strings.ToUpper(strings.ReplaceAll(opts.EnvKey, ".", "_"))
		envName := strings.ToLower(os.Getenv(envVarName))
		if envName == "" {
			envName = k.String(opts.EnvKey)
		}
		if envName != "" {
			overlayFile := filepath.Join(opts.BaseDir, "config."+envName+".yaml")
			if err := k.Load(file.Provider(overlayFile), koanyaml.Parser()); err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					return fmt.Errorf("config: load %s: %w", overlayFile, err)
				}
			}
		}
	}

	// 3. Environment variables — override YAML after the schema is populated,
	// so transformKey can call k.Exists() to resolve ambiguous underscores.
	if opts.EnvPrefix != "" {
		envProvider := env.Provider(opts.EnvPrefix, ".", func(s string) string {
			rawKey := strings.ToLower(strings.TrimPrefix(s, opts.EnvPrefix))
			return transformKey(k, rawKey)
		})
		if err := k.Load(envProvider, nil); err != nil {
			return fmt.Errorf("config: load env vars: %w", err)
		}
	}

	// 4. Content files — loaded after env vars so content values cannot be
	// accidentally overridden by deployment environment variables.
	for _, cf := range opts.ContentFiles {
		if cf.Filename == "" {
			continue
		}
		path := filepath.Join(opts.BaseDir, cf.Filename)
		_ = k.Load(file.Provider(path), koanyaml.Parser()) // missing = fine
	}

	// 5. Unmarshal merged config state into the caller's target struct.
	if err := k.Unmarshal("", target); err != nil {
		return fmt.Errorf("config: unmarshal: %w", err)
	}

	// 6. Per-file validation: each ContentFile with a non-nil Target gets its
	// own validation pass so errors are attributed to the specific file.
	v := validator.New()
	if opts.ValidatorSetup != nil {
		opts.ValidatorSetup(v)
	}
	for _, cf := range opts.ContentFiles {
		if cf.Target == nil {
			continue
		}
		if err := v.Struct(cf.Target); err != nil {
			return fmt.Errorf("config: %s: %w", cf.Filename, err)
		}
	}

	// 7. Validate the whole config using struct tags (required, min/max, oneof, etc.).
	if err := v.Struct(target); err != nil {
		return fmt.Errorf("config: validation: %w", err)
	}

	return nil
}

// Load allocates a fresh *T, calls opts(cfg) to build the Options (which
// allows ContentFile.Target to hold addresses of fields within the freshly
// allocated struct), then delegates to loadConfig.
func Load[T any](opts func(*T) Options) (*T, error) {
	cfg := new(T)
	if err := loadConfig(cfg, opts(cfg)); err != nil {
		return nil, err
	}
	return cfg, nil
}

// transformKey converts a lowercased, prefix-stripped env var name into the
// matching koanf key path. It tries candidate forms from most-dots to
// most-underscores, returning the first form where k.Exists() is true.
//
// This disambiguates level separators from underscores in field names without
// any hardcoded knowledge of the config schema.
//
// Example for "auth_session_auth_key" (from CTX_AUTH_SESSION_AUTH_KEY):
//
//	n=0: "auth.session.auth.key"   → k.Exists? no
//	n=1: "auth.session.auth_key"   → k.Exists? yes → return
//
// Example for "server_graceful_timeout" (from CTX_SERVER_GRACEFUL_TIMEOUT):
//
//	n=0: "server.graceful.timeout" → k.Exists? no
//	n=1: "server.graceful_timeout" → k.Exists? yes → return
func transformKey(k *koanf.Koanf, rawKey string) string {
	parts := strings.Split(rawKey, "_")

	for n := range parts {
		pivot := len(parts) - n

		var candidate string
		if pivot == len(parts) {
			// n==0: all separators become dots
			candidate = strings.Join(parts, ".")
		} else {
			left := strings.Join(parts[:pivot], ".")
			right := strings.Join(parts[pivot:], "_")
			candidate = left + "_" + right
		}

		if k.Exists(candidate) {
			return candidate
		}
	}

	// No match found — return the all-dots form. koanf will store it at an
	// unknown path that won't map to any struct field (safely ignored).
	return strings.Join(parts, ".")
}
