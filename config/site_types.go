package config

type SiteConfig struct {
	Title         string   `koanf:"title"          validate:"required"`
	Description   string   `koanf:"description"`
	LogoPath      string   `koanf:"logo_path"`
	FaviconPath   string   `koanf:"favicon_path"`
	MetaKeywords  []string `koanf:"meta_keywords"`
	OGImage       string   `koanf:"og_image"`
	CopyrightYear int      `koanf:"copyright_year"`
}
