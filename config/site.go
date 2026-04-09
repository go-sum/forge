package config

import (
	"github.com/go-sum/componentry/patterns/font"
	"github.com/go-sum/site"
)

// SiteConfig holds site presentation metadata.
type SiteConfig struct {
	Title         string `validate:"required"`
	Description   string
	LogoPath      string
	FaviconPath   string
	MetaKeywords  []string
	OGImage       string
	CopyrightYear int
	Robots        site.RobotsConfig
	Sitemap       site.SitemapConfig
	Fonts         font.Config
}

func defaultSite() SiteConfig {
	return SiteConfig{
		Title:         "starter",
		Description:   "",
		LogoPath:      "",
		FaviconPath:   "",
		MetaKeywords:  []string{},
		OGImage:       "",
		CopyrightYear: 2025,
		Robots: site.RobotsConfig{
			DefaultAllow: true,
			DisallowPaths: []string{
				"/_components",
				"/admin",
				"/profile",
				"/signin",
				"/signup",
				"/health",
			},
			CacheControl: "public, max-age=86400",
		},
		Sitemap: site.SitemapConfig{
			DefaultChangeFreq: "weekly",
			DefaultPriority:   0.5,
			Routes: []site.RouteEntry{
				{Name: "home.show", ChangeFreq: "daily", Priority: floatPtr(1.0)},
			},
			StaticPages:  []site.StaticEntry{},
			CacheControl: "public, max-age=3600",
		},
		Fonts: font.Config{
			Google:     font.GoogleConfig{Families: []string{}},
			Bunny:      font.BunnyConfig{Families: []string{}},
			Adobe:      font.AdobeConfig{ProjectID: ""},
			SelfHosted: []font.SelfHostedGroup{},
		},
	}
}

func floatPtr(f float64) *float64 { return &f }
