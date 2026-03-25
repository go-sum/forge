package site_test

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/go-sum/site"
)

// xmlURLSetParsed is used only for round-trip verification.
type xmlURLSetParsed struct {
	XMLName xml.Name         `xml:"urlset"`
	XMLNS   string           `xml:"xmlns,attr"`
	URLs    []xmlURLParsed   `xml:"url"`
}

type xmlURLParsed struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

// fp returns a pointer to a float64 value for use in Entry.Priority.
func fp(v float64) *float64 { return &v }

func TestBuildSitemap(t *testing.T) {
	modTime := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		entries   []site.Entry
		wantLines []string
		noLines   []string
	}{
		{
			name:      "empty_entries_produces_valid_urlset",
			entries:   nil,
			wantLines: []string{`<?xml`, `<urlset`, `xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"`},
			noLines:   []string{"<url>"},
		},
		{
			name: "single_entry_all_fields",
			entries: []site.Entry{
				{
					Loc:        "https://example.com/",
					LastMod:    &modTime,
					ChangeFreq: "daily",
					Priority:   fp(1.0),
				},
			},
			wantLines: []string{
				"<loc>https://example.com/</loc>",
				"<lastmod>2026-03-24</lastmod>",
				"<changefreq>daily</changefreq>",
				"<priority>1.0</priority>",
			},
		},
		{
			name: "nil_priority_omits_element",
			entries: []site.Entry{
				{Loc: "https://example.com/about", Priority: nil},
			},
			noLines: []string{"<priority>"},
		},
		{
			name: "explicit_zero_priority_emitted",
			entries: []site.Entry{
				{Loc: "https://example.com/low", Priority: fp(0.0)},
			},
			wantLines: []string{"<priority>0.0</priority>"},
		},
		{
			name: "lastmod_nil_omits_element",
			entries: []site.Entry{
				{Loc: "https://example.com/contact", LastMod: nil},
			},
			noLines: []string{"<lastmod>"},
		},
		{
			name: "multiple_entries",
			entries: []site.Entry{
				{Loc: "https://example.com/", Priority: fp(1.0), ChangeFreq: "daily"},
				{Loc: "https://example.com/about", Priority: fp(0.8), ChangeFreq: "monthly"},
			},
			wantLines: []string{
				"<loc>https://example.com/</loc>",
				"<loc>https://example.com/about</loc>",
				"<priority>0.8</priority>",
			},
		},
		{
			name: "xml_declaration_present",
			entries: []site.Entry{
				{Loc: "https://example.com/"},
			},
			wantLines: []string{`<?xml version="1.0" encoding="UTF-8"?>`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := site.BuildSitemap(tc.entries)
			if err != nil {
				t.Fatalf("BuildSitemap() error = %v", err)
			}
			body := string(got)
			for _, want := range tc.wantLines {
				if !strings.Contains(body, want) {
					t.Errorf("output missing %q\ngot:\n%s", want, body)
				}
			}
			for _, absent := range tc.noLines {
				if strings.Contains(body, absent) {
					t.Errorf("output should not contain %q\ngot:\n%s", absent, body)
				}
			}
		})
	}
}

func TestBuildSitemapXMLRoundTrip(t *testing.T) {
	modTime := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)
	entries := []site.Entry{
		{
			Loc:        "https://example.com/",
			LastMod:    &modTime,
			ChangeFreq: "daily",
			Priority:   fp(1.0),
		},
		{
			Loc:        "https://example.com/about",
			ChangeFreq: "monthly",
			Priority:   fp(0.5),
		},
	}

	data, err := site.BuildSitemap(entries)
	if err != nil {
		t.Fatalf("BuildSitemap() error = %v", err)
	}

	var parsed xmlURLSetParsed
	if err := xml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("xml.Unmarshal() error = %v\noutput:\n%s", err, string(data))
	}

	if parsed.XMLNS != "http://www.sitemaps.org/schemas/sitemap/0.9" {
		t.Errorf("xmlns = %q, want sitemap namespace", parsed.XMLNS)
	}
	if len(parsed.URLs) != 2 {
		t.Fatalf("url count = %d, want 2", len(parsed.URLs))
	}
	if parsed.URLs[0].Loc != "https://example.com/" {
		t.Errorf("URLs[0].Loc = %q, want %q", parsed.URLs[0].Loc, "https://example.com/")
	}
	if parsed.URLs[0].LastMod != "2026-03-24" {
		t.Errorf("URLs[0].LastMod = %q, want 2026-03-24", parsed.URLs[0].LastMod)
	}
	if parsed.URLs[0].ChangeFreq != "daily" {
		t.Errorf("URLs[0].ChangeFreq = %q, want daily", parsed.URLs[0].ChangeFreq)
	}
	if parsed.URLs[0].Priority != "1.0" {
		t.Errorf("URLs[0].Priority = %q, want 1.0", parsed.URLs[0].Priority)
	}
}
