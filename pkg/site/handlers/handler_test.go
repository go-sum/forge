package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sum/site"
	"github.com/go-sum/site/handlers"
	"github.com/labstack/echo/v5"
)

// fp returns a pointer to a float64 for use in sitemap Priority fields.
func fp(v float64) *float64 { return &v }

func newTestRoutes() func() echo.Routes {
	e := echo.New()
	noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }
	mustAdd := func(r echo.Route) {
		if _, err := e.AddRoute(r); err != nil {
			panic(err)
		}
	}
	mustAdd(echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: noOp})
	mustAdd(echo.Route{Method: http.MethodGet, Path: "/users/:id/edit", Name: "user.edit", Handler: noOp})
	return func() echo.Routes { return e.Router().Routes() }
}

func newTestHandler(origin string, robotsCfg site.RobotsConfig, sitemapCfg site.SitemapConfig) *handlers.Handler {
	return handlers.New(handlers.Config{
		Origin:  origin,
		Robots:  robotsCfg,
		Sitemap: sitemapCfg,
	}, newTestRoutes())
}

func newRequestContext(method, target string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// TestRobotsTxt exercises the /robots.txt handler.
func TestRobotsTxt(t *testing.T) {
	tests := []struct {
		name      string
		robots    site.RobotsConfig
		wantLines []string
		noLines   []string
		wantCC    string
	}{
		{
			name:   "default_allow_includes_disallow_paths",
			robots: site.RobotsConfig{DefaultAllow: true},
			wantLines: []string{
				"User-agent: *",
				"Disallow: /signin",
				"Sitemap: https://example.com/sitemap.xml",
			},
			wantCC: "public, max-age=86400",
		},
		{
			name:   "disallow_all_when_default_allow_false",
			robots: site.RobotsConfig{DefaultAllow: false},
			wantLines: []string{
				"User-agent: *",
				"Disallow: /",
			},
			wantCC: "public, max-age=86400",
		},
		{
			name: "custom_disallow_paths_override_defaults",
			robots: site.RobotsConfig{
				DefaultAllow:  true,
				DisallowPaths: []string{"/private"},
			},
			wantLines: []string{"Disallow: /private"},
			noLines:   []string{"Disallow: /signin"},
			wantCC:    "public, max-age=86400",
		},
		{
			name:      "sitemap_url_derived_from_origin",
			robots:    site.RobotsConfig{DefaultAllow: true},
			wantLines: []string{"Sitemap: https://example.com/sitemap.xml"},
			wantCC:    "public, max-age=86400",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler("https://example.com", tc.robots, site.SitemapConfig{})
			c, rec := newRequestContext(http.MethodGet, "/robots.txt")

			if err := h.RobotsTxt(c); err != nil {
				t.Fatalf("RobotsTxt() error = %v", err)
			}
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			ct := rec.Header().Get("Content-Type")
			if !strings.HasPrefix(ct, "text/plain") {
				t.Errorf("Content-Type = %q, want text/plain", ct)
			}
			if got := rec.Header().Get("Cache-Control"); got != tc.wantCC {
				t.Errorf("Cache-Control = %q, want %q", got, tc.wantCC)
			}
			body := rec.Body.String()
			for _, want := range tc.wantLines {
				if !strings.Contains(body, want) {
					t.Errorf("body missing %q\nbody:\n%s", want, body)
				}
			}
			for _, absent := range tc.noLines {
				if strings.Contains(body, absent) {
					t.Errorf("body should not contain %q\nbody:\n%s", absent, body)
				}
			}
		})
	}
}

// TestSitemapXML exercises the /sitemap.xml handler.
func TestSitemapXML(t *testing.T) {
	tests := []struct {
		name      string
		sitemap   site.SitemapConfig
		wantLines []string
		noLines   []string
		wantCC    string
	}{
		{
			name:      "empty_config_produces_valid_urlset",
			sitemap:   site.SitemapConfig{},
			wantLines: []string{`<?xml`, `<urlset`},
			noLines:   []string{"<url>"},
			wantCC:    "public, max-age=3600",
		},
		{
			name: "named_route_resolved_to_absolute_url",
			sitemap: site.SitemapConfig{
				Routes: []site.RouteEntry{
					{Name: "home.show", ChangeFreq: "daily", Priority: fp(1.0)},
				},
			},
			wantLines: []string{
				"<loc>https://example.com/</loc>",
				"<changefreq>daily</changefreq>",
				"<priority>1.0</priority>",
			},
			wantCC: "public, max-age=3600",
		},
		{
			name: "parameterized_route_is_skipped",
			sitemap: site.SitemapConfig{
				Routes: []site.RouteEntry{
					{Name: "user.edit", Priority: fp(0.8)},
				},
			},
			// user.edit resolves to /users/:id/edit — must not appear
			noLines: []string{"/users/"},
			wantCC:  "public, max-age=3600",
		},
		{
			name: "unknown_route_name_is_skipped",
			sitemap: site.SitemapConfig{
				Routes: []site.RouteEntry{
					{Name: "nonexistent.route"},
				},
			},
			wantLines: []string{`<urlset`},
			noLines:   []string{"nonexistent"},
			wantCC:    "public, max-age=3600",
		},
		{
			name: "static_page_entry_included",
			sitemap: site.SitemapConfig{
				StaticPages: []site.StaticEntry{
					{Path: "/about", ChangeFreq: "monthly", Priority: fp(0.5)},
				},
			},
			wantLines: []string{
				"<loc>https://example.com/about</loc>",
				"<changefreq>monthly</changefreq>",
			},
			wantCC: "public, max-age=3600",
		},
		{
			name: "default_changefreq_and_priority_applied",
			sitemap: site.SitemapConfig{
				DefaultChangeFreq: "weekly",
				DefaultPriority:   0.6,
				Routes: []site.RouteEntry{
					{Name: "home.show"}, // no explicit changefreq/priority
				},
			},
			wantLines: []string{
				"<changefreq>weekly</changefreq>",
				"<priority>0.6</priority>",
			},
			wantCC: "public, max-age=3600",
		},
		{
			name: "explicit_zero_priority_emitted_not_replaced_by_default",
			sitemap: site.SitemapConfig{
				DefaultPriority: 0.5,
				Routes: []site.RouteEntry{
					{Name: "home.show", Priority: fp(0.0)},
				},
			},
			wantLines: []string{"<priority>0.0</priority>"},
			wantCC:    "public, max-age=3600",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler("https://example.com", site.RobotsConfig{}, tc.sitemap)
			c, rec := newRequestContext(http.MethodGet, "/sitemap.xml")

			if err := h.SitemapXML(c); err != nil {
				t.Fatalf("SitemapXML() error = %v", err)
			}
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			ct := rec.Header().Get("Content-Type")
			if !strings.HasPrefix(ct, "application/xml") {
				t.Errorf("Content-Type = %q, want application/xml", ct)
			}
			if got := rec.Header().Get("Cache-Control"); got != tc.wantCC {
				t.Errorf("Cache-Control = %q, want %q", got, tc.wantCC)
			}
			body := rec.Body.String()
			for _, want := range tc.wantLines {
				if !strings.Contains(body, want) {
					t.Errorf("body missing %q\nbody:\n%s", want, body)
				}
			}
			for _, absent := range tc.noLines {
				if strings.Contains(body, absent) {
					t.Errorf("body should not contain %q\nbody:\n%s", absent, body)
				}
			}
		})
	}
}
