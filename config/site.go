package config

import (
	"github.com/go-sum/componentry/patterns/font"
	"github.com/go-sum/site"
)

// SiteConfig holds site presentation metadata loaded from config/site.yaml.
type SiteConfig struct {
	Title         string             `koanf:"title"	validate:"required"`
	Description   string             `koanf:"description"`
	LogoPath      string             `koanf:"logo_path"`
	FaviconPath   string             `koanf:"favicon_path"`
	MetaKeywords  []string           `koanf:"meta_keywords"`
	OGImage       string             `koanf:"og_image"`
	CopyrightYear int                `koanf:"copyright_year"`
	Robots        site.RobotsConfig  `koanf:"robots"`
	Sitemap       site.SitemapConfig `koanf:"sitemap"`
	Fonts         font.Config        `koanf:"fonts"`
}
