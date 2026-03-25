// Package handlers provides HTTP handlers for serving /robots.txt and
// /sitemap.xml. It wraps the site generation functions from the parent
// package and adds HTTP transport concerns (caching, content types).
package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-sum/site"
	"github.com/labstack/echo/v5"
)

const (
	robotsCacheControl  = "public, max-age=86400"
	sitemapCacheControl = "public, max-age=3600"
)

// Config holds everything the site HTTP handlers need.
type Config struct {
	// Origin is the external canonical origin, e.g. "https://example.com".
	// Used to build the Sitemap URL in robots.txt and absolute <loc> values
	// in sitemap.xml.
	Origin  string
	Robots  site.RobotsConfig
	Sitemap site.SitemapConfig
}

// Handler serves /robots.txt and /sitemap.xml.
type Handler struct {
	cfg    Config
	routes func() echo.Routes
}

// New constructs a Handler from cfg and a lazy route accessor.
// routes is evaluated at request time so it reflects all registered routes.
func New(cfg Config, routes func() echo.Routes) *Handler {
	return &Handler{cfg: cfg, routes: routes}
}

// RobotsTxt generates and serves /robots.txt.
// The sitemap URL is derived from Config.Origin.
// Cache-Control: public, max-age=86400 (24 hours).
func (h *Handler) RobotsTxt(c *echo.Context) error {
	robotsCfg := h.cfg.Robots
	robotsCfg.SitemapURL = h.cfg.Origin + "/sitemap.xml"

	content, err := site.BuildRobots(robotsCfg)
	if err != nil {
		return fmt.Errorf("robots.txt: %w", err)
	}

	c.Response().Header().Set("Cache-Control", robotsCacheControl)
	return c.String(http.StatusOK, content)
}

// SitemapXML generates and serves /sitemap.xml from named routes and
// static pages declared in Config.Sitemap.
// Cache-Control: public, max-age=3600 (1 hour).
func (h *Handler) SitemapXML(c *echo.Context) error {
	entries := h.buildSitemapEntries()

	data, err := site.BuildSitemap(entries)
	if err != nil {
		return fmt.Errorf("sitemap.xml: %w", err)
	}

	c.Response().Header().Set("Cache-Control", sitemapCacheControl)
	return c.Blob(http.StatusOK, "application/xml; charset=utf-8", data)
}

// buildSitemapEntries assembles the full entry list from config:
//  1. Named routes resolved via safeReverse (parameterized routes skipped).
//  2. Static pages with explicit paths prepended with Origin.
//
// Per-entry changefreq and priority fall back to SitemapConfig defaults.
func (h *Handler) buildSitemapEntries() []site.Entry {
	cfg := h.cfg.Sitemap
	origin := h.cfg.Origin
	routes := h.routes()

	var entries []site.Entry

	for _, r := range cfg.Routes {
		path, ok := safeReverse(routes, r.Name)
		if !ok {
			continue
		}

		changefreq := r.ChangeFreq
		if changefreq == "" {
			changefreq = cfg.DefaultChangeFreq
		}

		entries = append(entries, site.Entry{
			Loc:        origin + path,
			ChangeFreq: changefreq,
			Priority:   resolvePriority(r.Priority, cfg.DefaultPriority),
		})
	}

	for _, sp := range cfg.StaticPages {
		changefreq := sp.ChangeFreq
		if changefreq == "" {
			changefreq = cfg.DefaultChangeFreq
		}

		entries = append(entries, site.Entry{
			Loc:        origin + sp.Path,
			ChangeFreq: changefreq,
			Priority:   resolvePriority(sp.Priority, cfg.DefaultPriority),
		})
	}

	return entries
}

// resolvePriority returns the entry's explicit priority when set, the default
// when non-zero, or nil (omit <priority> from XML) when both are unset.
func resolvePriority(entry *float64, defaultPriority float64) *float64 {
	if entry != nil {
		return entry
	}
	if defaultPriority != 0 {
		return &defaultPriority
	}
	return nil
}

// safeReverse resolves a named route to its path without panicking.
// Returns ("", false) if the route name is unknown or the resolved path
// still contains ":" — indicating an unfilled path parameter (e.g.
// /users/:id/edit). Such routes produce invalid sitemap URLs and are skipped.
func safeReverse(routes echo.Routes, name string) (string, bool) {
	path, err := routes.Reverse(name)
	if err != nil {
		return "", false
	}
	if strings.Contains(path, ":") {
		return "", false
	}
	return path, true
}
