package site

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

const sitemapXMLNS = "http://www.sitemaps.org/schemas/sitemap/0.9"

// Entry represents a single URL entry in a sitemap.
type Entry struct {
	// Loc is the absolute URL of the page (required).
	Loc string

	// LastMod is the optional last modification date.
	// Formatted as YYYY-MM-DD in the output.
	LastMod *time.Time

	// ChangeFreq is the optional change frequency hint.
	// Valid values: always, hourly, daily, weekly, monthly, yearly, never.
	ChangeFreq string

	// Priority is the optional relative priority (0.0–1.0).
	// A nil value omits the <priority> element (the crawler assumes 0.5 by default).
	// Use a non-nil pointer to explicitly emit any value, including 0.0.
	Priority *float64
}

// xmlURL is the wire encoding for a single <url> element.
type xmlURL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod,omitempty"`
	ChangeFreq string   `xml:"changefreq,omitempty"`
	Priority   string   `xml:"priority,omitempty"`
}

// xmlURLSet is the root <urlset> element.
type xmlURLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []xmlURL
}

// BuildSitemap generates a sitemap.xml document from entries.
// The returned bytes include the XML declaration header.
// An empty entries slice returns a valid but empty <urlset>.
func BuildSitemap(entries []Entry) ([]byte, error) {
	set := xmlURLSet{
		XMLNS: sitemapXMLNS,
		URLs:  make([]xmlURL, 0, len(entries)),
	}

	for _, e := range entries {
		u := xmlURL{
			Loc:        e.Loc,
			ChangeFreq: e.ChangeFreq,
		}
		if e.LastMod != nil {
			u.LastMod = e.LastMod.Format("2006-01-02")
		}
		if e.Priority != nil {
			u.Priority = fmt.Sprintf("%.1f", *e.Priority)
		}
		set.URLs = append(set.URLs, u)
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")

	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(set); err != nil {
		return nil, fmt.Errorf("site: encode sitemap: %w", err)
	}

	return buf.Bytes(), nil
}

// SitemapConfig holds the sitemap.xml generation settings.
// It is used as the sitemap section of the application site config
// (site.yaml), deserialized via the koanf tags.
type SitemapConfig struct {
	// Routes lists named application routes to include.
	// Each Name must match a registered route name resolvable without path
	// parameters (parameterized routes are silently skipped).
	Routes []RouteEntry `koanf:"routes" validate:"omitempty,dive"`

	// StaticPages lists explicit path entries to include in the sitemap.
	// Each Path is an absolute path; the handler prepends ExternalOrigin.
	StaticPages []StaticEntry `koanf:"static_pages" validate:"omitempty,dive"`

	// DefaultChangeFreq is the change frequency applied to entries that
	// do not specify one. Valid values: always, hourly, daily, weekly,
	// monthly, yearly, never.
	DefaultChangeFreq string `koanf:"default_changefreq" validate:"omitempty,oneof=always hourly daily weekly monthly yearly never"`

	// DefaultPriority is the priority (0.0–1.0) applied to entries that
	// specify a zero value.
	DefaultPriority float64 `koanf:"default_priority" validate:"omitempty,min=0,max=1"`
}

// RouteEntry configures a named application route for sitemap inclusion.
type RouteEntry struct {
	// Name is the registered route name (e.g. "home.show").
	Name string `koanf:"name" validate:"required"`

	// ChangeFreq overrides SitemapConfig.DefaultChangeFreq for this entry.
	ChangeFreq string `koanf:"changefreq" validate:"omitempty,oneof=always hourly daily weekly monthly yearly never"`

	// Priority overrides SitemapConfig.DefaultPriority for this entry.
	// Use a pointer so that 0.0 can be expressed explicitly (nil means "use default").
	Priority *float64 `koanf:"priority" validate:"omitempty,min=0,max=1"`
}

// StaticEntry configures an explicit static path for sitemap inclusion.
type StaticEntry struct {
	// Path is an absolute path (e.g. /about). The handler prepends ExternalOrigin.
	Path string `koanf:"path" validate:"required"`

	// ChangeFreq overrides SitemapConfig.DefaultChangeFreq for this entry.
	ChangeFreq string `koanf:"changefreq" validate:"omitempty,oneof=always hourly daily weekly monthly yearly never"`

	// Priority overrides SitemapConfig.DefaultPriority for this entry.
	// Use a pointer so that 0.0 can be expressed explicitly (nil means "use default").
	Priority *float64 `koanf:"priority" validate:"omitempty,min=0,max=1"`
}
