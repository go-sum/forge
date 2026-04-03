// Package site provides utilities for generating site metadata files
// such as robots.txt and sitemap.xml.
//
// This package has no external dependencies and can be used standalone
// in any Go module.
package site

import (
	"fmt"
	"strings"
)

// DefaultDisallowPaths is the list of paths disallowed when DefaultAllow is
// true and DisallowPaths is empty. These are internal surfaces or auth
// mutation endpoints that have no SEO value.
var DefaultDisallowPaths = []string{
	"/_components",
	"/account",
	"/users",
	"/signin",
	"/signup",
	"/signout",
	"/health",
}

// RobotsConfig controls what BuildRobots emits.
// It is also used as the robots section of the application site config
// (site.yaml), deserialized via the koanf tags.
type RobotsConfig struct {
	// DefaultAllow true means all crawlers are allowed by default; specific
	// paths in DisallowPaths are excluded.
	// DefaultAllow false emits "Disallow: /" to block all crawlers.
	DefaultAllow bool `koanf:"default_allow"`

	// DisallowPaths is the list of paths to disallow when DefaultAllow is
	// true. When nil or empty, DefaultDisallowPaths is used instead.
	DisallowPaths []string `koanf:"disallow_paths"`

	// SitemapURL is the absolute URL of the sitemap
	// (e.g. https://example.com/sitemap.xml). When non-empty, a
	// "Sitemap:" directive is appended to the output.
	// This field is derived at handler time from the application origin and
	// is not populated from site.yaml.
	SitemapURL string `koanf:"-"`
}

// BuildRobots generates a robots.txt document from cfg.
// The output is always valid robots.txt — an empty config produces a
// permissive file that allows all crawlers without any disallow rules.
func BuildRobots(cfg RobotsConfig) (string, error) {
	var b strings.Builder
	b.WriteString("User-agent: *\n")

	if !cfg.DefaultAllow {
		b.WriteString("Disallow: /\n")
	} else {
		paths := cfg.DisallowPaths
		if len(paths) == 0 {
			paths = DefaultDisallowPaths
		}
		for _, p := range paths {
			fmt.Fprintf(&b, "Disallow: %s\n", p)
		}
	}

	if cfg.SitemapURL != "" {
		fmt.Fprintf(&b, "\nSitemap: %s\n", cfg.SitemapURL)
	}

	return b.String(), nil
}
