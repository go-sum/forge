// Package font provides Gomponents helpers for loading web fonts in <head>.
//
// It supports three remote providers and one self-hosted strategy:
//
//   - [Google]: Google Fonts via preconnect + stylesheet link
//   - [Bunny]: Bunny Fonts (privacy-friendly Google Fonts alternative)
//   - [Adobe]: Adobe Fonts (Typekit) via stylesheet link
//   - [Self]: Self-hosted fonts via preload hints + @font-face declarations
//
// All provider types implement [Provider] so callers can collect them with [Nodes]:
//
//	nodes := font.Nodes(
//	    font.Google("Inter:wght@400;500;600;700"),
//	    font.Self(font.Face{
//	        Family: "MyFont",
//	        URL:    "/public/fonts/myfont-regular.woff2",
//	        Format: "woff2",
//	        Weight: "400",
//	        Style:  "normal",
//	    }),
//	)
//
// The returned []g.Node slice is suitable for inclusion in head.Props.Extra.
package font

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Provider is the common interface for all font sources.
// Each provider type returns a slice of <head> nodes (preconnect links,
// stylesheet links, preload hints, or inline style blocks), and declares
// the CSP source additions it requires.
type Provider interface {
	Nodes() []g.Node
	// CSPSources returns the Content-Security-Policy additions this provider
	// requires. Returns a zero-value CSPSources when the provider is empty/disabled.
	CSPSources() CSPSources
}

// CSPSources describes the Content-Security-Policy additions a font provider
// requires. All fields are additive — callers append them to existing CSP directives.
type CSPSources struct {
	// StyleSources lists external domains to allow in the style-src directive,
	// e.g. "https://fonts.googleapis.com".
	StyleSources []string
	// FontSources lists external domains to allow in the font-src directive,
	// e.g. "https://fonts.gstatic.com".
	FontSources []string
	// StyleInlineHashes lists 'sha256-…' hashes for inline <style> blocks that
	// must be permitted in the style-src directive.
	StyleInlineHashes []string
}

// Nodes collects and returns all <head> nodes from the given providers.
// The returned slice is nil when no providers are supplied.
func Nodes(providers ...Provider) []g.Node {
	var all []g.Node
	for _, p := range providers {
		all = append(all, p.Nodes()...)
	}
	return all
}

// CollectCSPSources merges the CSPSources from all providers into a single result.
func CollectCSPSources(providers []Provider) CSPSources {
	var result CSPSources
	for _, p := range providers {
		srcs := p.CSPSources()
		result.StyleSources = append(result.StyleSources, srcs.StyleSources...)
		result.FontSources = append(result.FontSources, srcs.FontSources...)
		result.StyleInlineHashes = append(result.StyleInlineHashes, srcs.StyleInlineHashes...)
	}
	return result
}

// --- Google Fonts ---

// GoogleProvider loads fonts from Google Fonts.
// It renders a preconnect pair to fonts.googleapis.com and fonts.gstatic.com,
// followed by a stylesheet link that requests all configured families.
type GoogleProvider struct {
	families []string
}

// Google returns a GoogleProvider for the given font family strings.
// Each family follows the Google Fonts v2 format, e.g. "Inter:wght@400;700".
// Multiple families are combined into a single stylesheet request.
func Google(families ...string) *GoogleProvider {
	return &GoogleProvider{families: families}
}

func (p *GoogleProvider) Nodes() []g.Node {
	if len(p.families) == 0 {
		return nil
	}
	href := "https://fonts.googleapis.com/css2?family=" +
		strings.Join(p.families, "&family=") +
		"&display=swap"
	return []g.Node{
		h.Link(h.Rel("preconnect"), h.Href("https://fonts.googleapis.com")),
		h.Link(h.Rel("preconnect"), h.Href("https://fonts.gstatic.com"), g.Attr("crossorigin", "")),
		h.Link(h.Rel("stylesheet"), h.Href(href)),
	}
}

func (p *GoogleProvider) CSPSources() CSPSources {
	if len(p.families) == 0 {
		return CSPSources{}
	}
	return CSPSources{
		StyleSources: []string{"https://fonts.googleapis.com"},
		FontSources:  []string{"https://fonts.gstatic.com"},
	}
}

// --- Bunny Fonts ---

// BunnyProvider loads fonts from Bunny Fonts (https://fonts.bunny.net),
// a GDPR-friendly alternative to Google Fonts with a compatible API.
// It renders a preconnect to fonts.bunny.net and a stylesheet link.
type BunnyProvider struct {
	families []string
}

// Bunny returns a BunnyProvider for the given font family strings.
// Each family follows the Bunny Fonts format, e.g. "inter:400,700".
func Bunny(families ...string) *BunnyProvider {
	return &BunnyProvider{families: families}
}

func (p *BunnyProvider) Nodes() []g.Node {
	if len(p.families) == 0 {
		return nil
	}
	href := "https://fonts.bunny.net/css?family=" +
		strings.Join(p.families, "&family=") +
		"&display=swap"
	return []g.Node{
		h.Link(h.Rel("preconnect"), h.Href("https://fonts.bunny.net")),
		h.Link(h.Rel("stylesheet"), h.Href(href)),
	}
}

func (p *BunnyProvider) CSPSources() CSPSources {
	if len(p.families) == 0 {
		return CSPSources{}
	}
	return CSPSources{
		StyleSources: []string{"https://fonts.bunny.net"},
		FontSources:  []string{"https://fonts.bunny.net"},
	}
}

// --- Adobe Fonts ---

// AdobeProvider loads fonts from Adobe Fonts (Typekit).
// It renders a single stylesheet link using the project's kit ID.
type AdobeProvider struct {
	projectID string
}

// Adobe returns an AdobeProvider for the given Typekit project ID.
// The project ID is visible in the Typekit embed code as the final path segment,
// e.g. for "https://use.typekit.net/abc1234.css" the project ID is "abc1234".
func Adobe(projectID string) *AdobeProvider {
	return &AdobeProvider{projectID: projectID}
}

func (p *AdobeProvider) Nodes() []g.Node {
	if p.projectID == "" {
		return nil
	}
	return []g.Node{
		h.Link(h.Rel("stylesheet"), h.Href("https://use.typekit.net/"+p.projectID+".css")),
	}
}

func (p *AdobeProvider) CSPSources() CSPSources {
	if p.projectID == "" {
		return CSPSources{}
	}
	return CSPSources{
		StyleSources: []string{"https://use.typekit.net"},
		FontSources:  []string{"https://use.typekit.net"},
	}
}

// --- Self-hosted Fonts ---

// Face describes one font file in a self-hosted font family.
type Face struct {
	// Family is the CSS font-family name, e.g. "Inter".
	Family string
	// URL is the fully-resolved public URL to the font file.
	// Use assets.Path("fonts/inter-regular.woff2") to get a content-hashed URL.
	URL string
	// Format is the font format hint: "woff2", "woff", or "truetype".
	Format string
	// Weight is the CSS font-weight value, e.g. "400", "700", or "normal".
	// Defaults to "400" when empty.
	Weight string
	// Style is the CSS font-style value: "normal" or "italic".
	// Defaults to "normal" when empty.
	Style string
}

// SelfHostedProvider renders preload hints and @font-face declarations for
// self-hosted font files. Preloading woff2 files eliminates render-blocking
// network round-trips for critical fonts.
type SelfHostedProvider struct {
	faces []Face
}

// Self returns a SelfHostedProvider for the given font faces.
// Each Face must supply at least Family and URL; Format, Weight, and Style
// default to "woff2", "400", and "normal" respectively when empty.
func Self(faces ...Face) *SelfHostedProvider {
	return &SelfHostedProvider{faces: faces}
}

func (p *SelfHostedProvider) Nodes() []g.Node {
	if len(p.faces) == 0 {
		return nil
	}

	nodes := make([]g.Node, 0, len(p.faces)+1)

	// Emit one <link rel="preload"> per woff2 face for early resource hints.
	for _, f := range p.faces {
		if f.URL == "" {
			continue
		}
		nodes = append(nodes,
			h.Link(
				h.Rel("preload"),
				h.Href(f.URL),
				g.Attr("as", "font"),
				g.Attr("type", "font/"+orDefault(f.Format, "woff2")),
				g.Attr("crossorigin", ""),
			),
		)
	}

	// Emit a single <style> block with one @font-face rule per face.
	nodes = append(nodes, fontFaceStyle(p.faces))
	return nodes
}

func (p *SelfHostedProvider) CSPSources() CSPSources {
	css := buildFontFaceCSS(p.faces)
	if css == "" {
		return CSPSources{}
	}
	sum := sha256.Sum256([]byte(css))
	hash := fmt.Sprintf("'sha256-%s'", base64.StdEncoding.EncodeToString(sum[:]))
	return CSPSources{
		StyleInlineHashes: []string{hash},
	}
}

// fontFaceStyle renders an inline <style> block containing one @font-face
// rule for each face. Inlining avoids an extra network request and keeps
// the critical font loading path entirely within the HTML document.
func fontFaceStyle(faces []Face) g.Node {
	css := buildFontFaceCSS(faces)
	if css == "" {
		return g.Text("")
	}
	return h.StyleEl(g.Raw(css))
}

// buildFontFaceCSS returns the raw CSS string for all @font-face rules.
// It is used both for rendering and for computing the CSP style-src hash.
func buildFontFaceCSS(faces []Face) string {
	var sb strings.Builder
	for _, f := range faces {
		if f.URL == "" || f.Family == "" {
			continue
		}
		format := orDefault(f.Format, "woff2")
		weight := orDefault(f.Weight, "400")
		style := orDefault(f.Style, "normal")

		sb.WriteString("@font-face{")
		sb.WriteString("font-family:'")
		sb.WriteString(f.Family)
		sb.WriteString("';")
		sb.WriteString("src:url('")
		sb.WriteString(f.URL)
		sb.WriteString("') format('")
		sb.WriteString(format)
		sb.WriteString("');")
		sb.WriteString("font-weight:")
		sb.WriteString(weight)
		sb.WriteString(";")
		sb.WriteString("font-style:")
		sb.WriteString(style)
		sb.WriteString(";")
		sb.WriteString("font-display:swap;")
		sb.WriteString("}")
	}
	return sb.String()
}

func orDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// --- Config types ---

// Config holds declarative font loading configuration.
// The koanf struct tags match the site.yaml fonts section so callers can
// unmarshal directly into this type without an intermediate config struct.
type Config struct {
	Google     GoogleConfig      `koanf:"google"`
	Bunny      BunnyConfig       `koanf:"bunny"`
	Adobe      AdobeConfig       `koanf:"adobe"`
	SelfHosted []SelfHostedGroup `koanf:"self_hosted"`
}

// GoogleConfig configures Google Fonts loading.
// Families use the Google Fonts v2 format, e.g. "Inter:wght@400;500;600;700".
type GoogleConfig struct {
	Families []string `koanf:"families"`
}

// BunnyConfig configures Bunny Fonts loading (GDPR-friendly Google Fonts alternative).
// Families use the Bunny Fonts format, e.g. "inter:400,700".
type BunnyConfig struct {
	Families []string `koanf:"families"`
}

// AdobeConfig configures Adobe Fonts (Typekit) loading via a kit project ID.
type AdobeConfig struct {
	ProjectID string `koanf:"project_id"`
}

// SelfHostedGroup describes a single font family with one or more face files.
type SelfHostedGroup struct {
	Family string            `koanf:"family" validate:"required"`
	Faces  []SelfHostedFace  `koanf:"faces"  validate:"required,min=1,dive"`
}

// SelfHostedFace describes one font file within a self-hosted family.
type SelfHostedFace struct {
	// URL is the public-relative path to the font file, e.g. "fonts/inter-400.woff2".
	// Pass it through a path resolver (e.g. assets.Path) to get a content-hashed URL.
	URL    string `koanf:"url"    validate:"required"`
	Format string `koanf:"format" validate:"omitempty,oneof=woff2 woff truetype"`
	Weight string `koanf:"weight"`
	Style  string `koanf:"style"  validate:"omitempty,oneof=normal italic"`
}

// BuildProviders converts a Config into a slice of Provider values ready to
// pass to Nodes or CollectCSPSources.
//
// pathFunc is called on each self-hosted font URL to resolve it to a public
// path, e.g. with a content hash. Pass assets.Path in production or
// func(s string) string { return s } in tests.
func BuildProviders(cfg Config, pathFunc func(string) string) []Provider {
	var providers []Provider
	if len(cfg.Google.Families) > 0 {
		providers = append(providers, Google(cfg.Google.Families...))
	}
	if len(cfg.Bunny.Families) > 0 {
		providers = append(providers, Bunny(cfg.Bunny.Families...))
	}
	if cfg.Adobe.ProjectID != "" {
		providers = append(providers, Adobe(cfg.Adobe.ProjectID))
	}
	for _, group := range cfg.SelfHosted {
		faces := make([]Face, 0, len(group.Faces))
		for _, f := range group.Faces {
			faces = append(faces, Face{
				Family: group.Family,
				URL:    pathFunc(f.URL),
				Format: f.Format,
				Weight: f.Weight,
				Style:  f.Style,
			})
		}
		if len(faces) > 0 {
			providers = append(providers, Self(faces...))
		}
	}
	return providers
}
