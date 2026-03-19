package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"starter/pkg/assetconfig"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// fetchSVG fetches SVG bytes from base+file. Supports http/https URLs,
// file:// URIs, and bare filesystem paths.
func fetchSVG(base, file string) ([]byte, error) {
	switch {
	case strings.HasPrefix(base, "http://") || strings.HasPrefix(base, "https://"):
		url := base + file
		resp, err := httpClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("GET %s: %w", url, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	case strings.HasPrefix(base, "file://"):
		path := filepath.Join(strings.TrimPrefix(base, "file://"), file)
		return os.ReadFile(path)
	default:
		return os.ReadFile(filepath.Join(base, file))
	}
}

var (
	reViewBox     = regexp.MustCompile("(?i)viewBox=[\"']([^\"']+)[\"']")
	reOuterSVG    = regexp.MustCompile("(?si)<svg[^>]*>(.*)</svg>")
	reOuterSVGTag = regexp.MustCompile("(?si)^<svg([^>]*)>")
	reScript      = regexp.MustCompile("(?si)<script[^>]*>.*?</script>")
	reEventAttr   = regexp.MustCompile("(?i)\\son\\w+=\"[^\"]*\"")
	rePresAttr    = regexp.MustCompile("(?i)\\b(fill|stroke|stroke-width|stroke-linecap|stroke-linejoin|stroke-dasharray|stroke-miterlimit|fill-rule|clip-rule)=\"([^\"]*)\"")
)

// processSVG extracts the viewBox and inner content from raw SVG bytes,
// sanitizes event handlers and scripts, and wraps in a <symbol> element.
// Presentation attributes (fill, stroke, etc.) are transferred from the outer
// <svg> tag to the <symbol> so icons render correctly via <use>.
func processSVG(data []byte, id string) (string, error) {
	s := string(data)

	viewBox := "0 0 24 24"
	if m := reViewBox.FindStringSubmatch(s); m != nil {
		viewBox = m[1]
	}

	var presAttrs string
	if m := reOuterSVGTag.FindStringSubmatch(s); m != nil {
		for _, match := range rePresAttr.FindAllStringSubmatch(m[1], -1) {
			presAttrs += fmt.Sprintf(" %s=\"%s\"", strings.ToLower(match[1]), match[2])
		}
	}

	inner := s
	if m := reOuterSVG.FindStringSubmatch(s); m != nil {
		inner = m[1]
	}

	inner = reScript.ReplaceAllString(inner, "")
	inner = reEventAttr.ReplaceAllString(inner, "")

	var lines []string
	for _, line := range strings.Split(inner, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, "      "+line)
		}
	}

	if len(lines) == 0 {
		return fmt.Sprintf("    <symbol id=\"%s\" viewBox=\"%s\"%s/>", id, viewBox, presAttrs), nil
	}

	return fmt.Sprintf("    <symbol id=\"%s\" viewBox=\"%s\"%s>\n%s\n    </symbol>", id, viewBox, presAttrs, strings.Join(lines, "\n")), nil
}

// buildSprite fetches all SVG sources concurrently and writes the assembled
// sprite file to cfg.Target (or prints it in dry-run mode).
//
// When every source is remote, the build is skipped if the target file already
// exists, avoiding unnecessary network fetches on every dev rebuild. Delete the
// target file to force a rebuild (or run `make sprites`). Any local source
// causes an unconditional rebuild since local files are fast reads and may have
// changed.
func buildSprite(name string, cfg assetconfig.SpriteConfig, dryRun bool) error {
	if !dryRun && allRemoteSources(cfg.Sources) {
		if _, err := os.Stat(cfg.Target); err == nil {
			fmt.Printf("  ↷ %s: target exists, skipping (delete to force rebuild)\n", name)
			return nil
		}
	}

	// Flatten all (source, file) pairs in declaration order to preserve symbol
	// ordering across sources while still fetching all files concurrently.
	type pair struct{ path, file string }
	var pairs []pair
	for _, src := range cfg.Sources {
		for _, file := range src.Files {
			pairs = append(pairs, pair{src.Path, file})
		}
	}

	symbols := make([]string, len(pairs))
	var eg errgroup.Group
	for i, p := range pairs {
		i, p := i, p
		eg.Go(func() error {
			data, err := fetchSVG(p.path, p.file)
			if err != nil {
				return fmt.Errorf("sprite %q, file %q: %w", name, p.file, err)
			}
			id := strings.TrimSuffix(p.file, ".svg")
			sym, err := processSVG(data, id)
			if err != nil {
				return fmt.Errorf("sprite %q, file %q: %w", name, p.file, err)
			}
			symbols[i] = sym
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	sb.WriteString("<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\">\n")
	sb.WriteString("  <defs>\n")
	for _, sym := range symbols {
		sb.WriteString(sym)
		sb.WriteString("\n")
	}
	sb.WriteString("  </defs>\n")
	sb.WriteString("</svg>\n")

	output := sb.String()
	if dryRun {
		fmt.Printf("--- [dry-run] %s -> %s (%d icons) ---\n", name, cfg.Target, len(pairs))
		fmt.Println(output)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Target), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(cfg.Target), err)
	}
	if err := os.WriteFile(cfg.Target, []byte(output), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", cfg.Target, err)
	}
	fmt.Printf("  ✓ %s -> %s (%d icons)\n", name, cfg.Target, len(pairs))
	return nil
}

// allRemoteSources reports whether every source in the slice uses an http/https
// URL. A sprite with any local source is always rebuilt.
func allRemoteSources(sources []assetconfig.SourcesConfig) bool {
	for _, src := range sources {
		if !strings.HasPrefix(src.Path, "http://") && !strings.HasPrefix(src.Path, "https://") {
			return false
		}
	}
	return len(sources) > 0
}

func buildSVGSprites(args []string) error {
	fs := flag.NewFlagSet("build-sprites", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := fs.String("config", assetconfig.DefaultConfigPath, "path to assets config file")
	spriteName := fs.String("sprite", "", "build only this named sprite (default: all enabled)")
	dryRun := fs.Bool("dry-run", false, "print output without writing files")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := assetconfig.Load(*configPath)
	if err != nil {
		return err
	}

	builtSprites, totalIcons := 0, 0
	for name, sprite := range cfg.Sprites {
		if !sprite.Enabled {
			continue
		}
		if *spriteName != "" && name != *spriteName {
			continue
		}
		if err := buildSprite(name, sprite, *dryRun); err != nil {
			return err
		}
		builtSprites++
		for _, src := range sprite.Sources {
			totalIcons += len(src.Files)
		}
	}

	fmt.Printf("Built %d sprite(s), %d total icons\n", builtSprites, totalIcons)
	return nil
}

func runBuildSVGSprites() {
	if err := buildSVGSprites(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
