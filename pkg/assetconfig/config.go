package assetconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	DefaultConfigPath   = ".assets.yaml"
	defaultSourceDir    = "static"
	defaultPublicDir    = "public"
	defaultPublicPrefix = "/public"
)

// Config is the top-level structure of .assets.yaml.
type Config struct {
	Paths   Paths                   `yaml:"paths"`
	JS      JSConfig                `yaml:"js"`
	CSS     []CSSConfig             `yaml:"css"`
	Sprites map[string]SpriteConfig `yaml:"sprites"`
}

// Paths defines the raw asset source tree plus the built public output.
type Paths struct {
	SourceDir    string `yaml:"source_dir"`
	PublicDir    string `yaml:"public_dir"`
	PublicPrefix string `yaml:"public_prefix"`
}

// JSConfig configures JavaScript asset downloading and syncing.
type JSConfig struct {
	Downloads []JSDownload `yaml:"downloads"`
	Sync      JSSyncConfig `yaml:"sync"`
}

// JSDownload describes a third-party JS file to fetch.
// {version} in URL is replaced at runtime with the resolved version.
type JSDownload struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	URL     string `yaml:"url"`
	Target  string `yaml:"target"`
}

// JSSyncConfig copies local JS source files into the public target directory,
// preserving files listed in Exclude (e.g. downloaded third-party assets).
type JSSyncConfig struct {
	Source  string   `yaml:"source"`
	Target  string   `yaml:"target"`
	Exclude []string `yaml:"exclude"`
}

// CSSConfig describes one CSS compilation step driven by an external tool.
type CSSConfig struct {
	Tool   string `yaml:"tool"`
	Input  string `yaml:"input"`
	Output string `yaml:"output"`
}

type SpriteConfig struct {
	Enabled bool            `yaml:"enabled"`
	Target  string          `yaml:"target"`
	Sources []SourcesConfig `yaml:"sources"`
}

type SourcesConfig struct {
	Path  string   `yaml:"path"`
	Files []string `yaml:"files"`
}

// Load reads .assets.yaml and normalizes source/public-relative paths against
// the configured roots so callers work with concrete filesystem locations.
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	cfg.Paths = cfg.Paths.withDefaults()
	cfg.normalize()
	return &cfg, nil
}

// SourceRoot returns the raw asset source directory.
func (p Paths) SourceRoot() string {
	return cleanPath(orDefault(p.SourceDir, defaultSourceDir))
}

// PublicRoot returns the built asset output directory served by the app.
func (p Paths) PublicRoot() string {
	return cleanPath(orDefault(p.PublicDir, defaultPublicDir))
}

// URLPrefix returns the URL prefix used to serve public assets.
func (p Paths) URLPrefix() string {
	prefix := strings.TrimSpace(p.PublicPrefix)
	if prefix == "" {
		prefix = defaultPublicPrefix
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if prefix != "/" {
		prefix = strings.TrimRight(prefix, "/")
	}
	return prefix
}

// PublicURL joins rel under the configured public URL prefix.
func (p Paths) PublicURL(rel string) string {
	rel = strings.TrimPrefix(filepath.ToSlash(rel), "/")
	if rel == "" {
		return p.URLPrefix()
	}
	return p.URLPrefix() + "/" + rel
}

func (p Paths) withDefaults() Paths {
	return Paths{
		SourceDir:    orDefault(p.SourceDir, defaultSourceDir),
		PublicDir:    orDefault(p.PublicDir, defaultPublicDir),
		PublicPrefix: orDefault(p.PublicPrefix, defaultPublicPrefix),
	}
}

func (c *Config) normalize() {
	for i := range c.JS.Downloads {
		c.JS.Downloads[i].Target = resolvePublicPath(c.Paths, c.JS.Downloads[i].Target)
	}

	c.JS.Sync.Source = resolveSourcePath(c.Paths, c.JS.Sync.Source)
	c.JS.Sync.Target = resolvePublicPath(c.Paths, c.JS.Sync.Target)

	for i := range c.CSS {
		c.CSS[i].Input = resolveSourcePath(c.Paths, c.CSS[i].Input)
		c.CSS[i].Output = resolvePublicPath(c.Paths, c.CSS[i].Output)
	}

	for name, sprite := range c.Sprites {
		sprite.Target = resolvePublicPath(c.Paths, sprite.Target)
		for i := range sprite.Sources {
			sprite.Sources[i].Path = resolveSpriteSourcePath(c.Paths, sprite.Sources[i].Path)
		}
		c.Sprites[name] = sprite
	}
}

func resolveSourcePath(paths Paths, value string) string {
	return resolveLocalPath(paths.SourceRoot(), value)
}

func resolvePublicPath(paths Paths, value string) string {
	return resolveLocalPath(paths.PublicRoot(), value)
}

func resolveSpriteSourcePath(paths Paths, value string) string {
	if isRemotePath(value) || strings.HasPrefix(value, "file://") {
		return value
	}
	return resolveLocalPath(paths.SourceRoot(), value)
}

func resolveLocalPath(root, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return root
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}

	cleanValue := filepath.Clean(value)
	if cleanValue == root || strings.HasPrefix(cleanValue, root+string(filepath.Separator)) {
		return cleanValue
	}
	return filepath.Clean(filepath.Join(root, cleanValue))
}

func cleanPath(path string) string {
	return filepath.Clean(strings.TrimSpace(path))
}

func orDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func isRemotePath(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}
