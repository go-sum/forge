// Package config provides a generic, reusable configuration loader built on
// koanf. It supports layered loading (YAML files), ${VAR} expansion via
// os.ExpandEnv in every YAML file, and struct validation via
// go-playground/validator tags.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
)

// ConfigFile is a configuration file entry.
// The first entry in Options.Files is the required base config; all others are optional.
type ConfigFile struct {
	// Filepath is the path to the YAML file (e.g. "config/app.yaml").
	// The first file in Options.Files is required; all others are silently skipped when absent.
	Filepath string
}

// Options configures how Load discovers and merges configuration sources.
type Options struct {
	// Files is the ordered list of configuration files to load.
	// The first file is required (error if missing); all subsequent files are optional
	// (silently skipped when absent). Last file wins on key conflicts.
	Files []ConfigFile

	// EnvKey is the active environment name (e.g. "development").
	// When non-empty, Load looks for an optional overlay file named
	// "{dir}/{stem}.{EnvKey}.yaml" alongside the first file.
	// Pass os.Getenv("APP_ENV") to drive this from the environment.
	EnvKey string
}

// loadConfig merges configuration from multiple sources into target and validates it.
// ${VAR} patterns in any YAML file are expanded via os.ExpandEnv before parsing.
//
// Loading order (last writer wins):
//  1. Files[0].Filepath                           — required; error if missing
//  2. {dir}/{stem}.{EnvKey}.yaml                  — optional overlay; silently skipped
//  3. Files[1:][*].Filepath                       — optional; silently skipped when absent
//  4. Unmarshal into target
//  5. Validate target using go-playground/validator struct tags
func loadConfig(target any, opts Options) error {
	if len(opts.Files) == 0 {
		return fmt.Errorf("config: no files specified")
	}

	k := koanf.New(".")

	// 1. First file is the required base config.
	base := opts.Files[0]
	if err := k.Load(&envExpandedFile{base.Filepath}, yaml.Parser()); err != nil {
		return fmt.Errorf("config: load %s: %w", base.Filepath, err)
	}

	// 2. Environment overlay (optional — silently skip if file is absent).
	if opts.EnvKey != "" {
		baseDir := filepath.Dir(base.Filepath)
		stem := strings.TrimSuffix(filepath.Base(base.Filepath), filepath.Ext(base.Filepath))
		overlayFile := filepath.Join(baseDir, stem+"."+opts.EnvKey+".yaml")
		if err := k.Load(&envExpandedFile{overlayFile}, yaml.Parser()); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("config: load %s: %w", overlayFile, err)
			}
		}
	}

	// 3. Remaining files are all optional.
	for _, cf := range opts.Files[1:] {
		if cf.Filepath == "" {
			continue
		}
		_ = k.Load(&envExpandedFile{cf.Filepath}, yaml.Parser()) // missing = fine
	}

	// 4. Unmarshal merged config state into the caller's target struct.
	if err := k.Unmarshal("", target); err != nil {
		return fmt.Errorf("config: unmarshal: %w", err)
	}

	// 5. Validate the whole config against its tagged structural schema.
	v := validator.New()
	if err := v.Struct(target); err != nil {
		return fmt.Errorf("config: validation: %w", err)
	}

	return nil
}

// Load allocates a fresh *T, calls opts(cfg) to build the Options, then
// delegates to loadConfig.
func Load[T any](opts func(*T) Options) (*T, error) {
	cfg := new(T)
	if err := loadConfig(cfg, opts(cfg)); err != nil {
		return nil, err
	}
	return cfg, nil
}

// envExpandedFile is a koanf Provider that reads a YAML file and expands
// ${VAR} patterns using os.ExpandEnv before returning the bytes for parsing.
// Unset variables expand to empty string.
type envExpandedFile struct{ path string }

func (e *envExpandedFile) ReadBytes() ([]byte, error) {
	b, err := os.ReadFile(e.path)
	if err != nil {
		return nil, err
	}
	return []byte(expandEnv(string(b))), nil
}

func (e *envExpandedFile) Read() (map[string]any, error) {
	return nil, errors.New("envExpandedFile does not support Read()")
}

// expandEnv replaces ${VAR} and ${VAR:-default} patterns in s using os.Getenv.
// Unset or empty variables use the default when the :- form is present;
// otherwise they expand to empty string.
func expandEnv(s string) string {
	var buf strings.Builder
	for {
		start := strings.Index(s, "${")
		if start == -1 {
			buf.WriteString(s)
			return buf.String()
		}
		buf.WriteString(s[:start])
		s = s[start+2:]
		end := strings.Index(s, "}")
		if end == -1 {
			// Unclosed placeholder — write literal and stop.
			buf.WriteString("${")
			buf.WriteString(s)
			return buf.String()
		}
		expr := s[:end]
		s = s[end+1:]
		if before, after, ok := strings.Cut(expr, ":-"); ok {
			key, def := before, after
			if v := os.Getenv(key); v != "" {
				buf.WriteString(v)
			} else {
				buf.WriteString(def)
			}
		} else {
			buf.WriteString(os.Getenv(expr))
		}
	}
}
