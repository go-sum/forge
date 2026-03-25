package site_test

import (
	"strings"
	"testing"

	"github.com/go-sum/site"
)

func TestBuildRobots(t *testing.T) {
	tests := []struct {
		name      string
		cfg       site.RobotsConfig
		wantLines []string
		noLines   []string
	}{
		{
			name: "disallow_all_when_default_allow_false",
			cfg:  site.RobotsConfig{DefaultAllow: false},
			wantLines: []string{
				"User-agent: *",
				"Disallow: /",
			},
			noLines: []string{"Disallow: /signin"},
		},
		{
			name: "default_allow_uses_package_defaults",
			cfg:  site.RobotsConfig{DefaultAllow: true},
			wantLines: []string{
				"User-agent: *",
				"Disallow: /signin",
				"Disallow: /signup",
				"Disallow: /signout",
				"Disallow: /_components",
				"Disallow: /users",
				"Disallow: /health",
			},
			noLines: []string{"Disallow: /\n"},
		},
		{
			name: "custom_disallow_overrides_defaults",
			cfg: site.RobotsConfig{
				DefaultAllow:  true,
				DisallowPaths: []string{"/custom", "/private"},
			},
			wantLines: []string{
				"User-agent: *",
				"Disallow: /custom",
				"Disallow: /private",
			},
			noLines: []string{"Disallow: /signin", "Disallow: /health"},
		},
		{
			name: "sitemap_directive_appended_when_url_set",
			cfg: site.RobotsConfig{
				DefaultAllow: true,
				SitemapURL:   "https://example.com/sitemap.xml",
			},
			wantLines: []string{
				"Sitemap: https://example.com/sitemap.xml",
			},
		},
		{
			name:    "no_sitemap_directive_when_url_empty",
			cfg:     site.RobotsConfig{DefaultAllow: true},
			noLines: []string{"Sitemap:"},
		},
		{
			name: "empty_config_is_permissive_with_default_disallows",
			cfg:  site.RobotsConfig{},
			// DefaultAllow=false → Disallow: /
			wantLines: []string{"User-agent: *", "Disallow: /"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := site.BuildRobots(tc.cfg)
			if err != nil {
				t.Fatalf("BuildRobots() error = %v", err)
			}
			for _, want := range tc.wantLines {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\ngot:\n%s", want, got)
				}
			}
			for _, absent := range tc.noLines {
				if strings.Contains(got, absent) {
					t.Errorf("output should not contain %q\ngot:\n%s", absent, got)
				}
			}
		})
	}
}
