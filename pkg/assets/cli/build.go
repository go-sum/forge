package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/go-sum/assets/config"
	"github.com/spf13/cobra"
)

type assetBuildOptions struct {
	ConfigPath string
	Minify     bool
	BuildCSS   bool
	BuildFonts bool
	BuildJS    bool
}

// resolveAssetsOpts builds the options struct from individual flag values.
// If none of css, js, fonts are true, all three default to true.
func resolveAssetsOpts(configPath string, minify, css, js, fonts bool) assetBuildOptions {
	if !css && !js && !fonts {
		css, js, fonts = true, true, true
	}
	return assetBuildOptions{
		ConfigPath: configPath,
		Minify:     minify,
		BuildCSS:   css,
		BuildJS:    js,
		BuildFonts: fonts,
	}
}

func newAssetsCmd() *cobra.Command {
	var configPath string
	var minify, css, js, fonts bool

	cmd := &cobra.Command{
		Use:   "assets",
		Short: "Build CSS, JS, and font assets",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := resolveAssetsOpts(configPath, minify, css, js, fonts)
			return buildAssets(opts)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", config.DefaultConfigPath, "path to assets config file")
	cmd.Flags().BoolVar(&minify, "minify", false, "minify compiled CSS and JS")
	cmd.Flags().BoolVar(&css, "css", false, "build CSS (default: all asset types)")
	cmd.Flags().BoolVar(&js, "js", false, "build JS (default: all asset types)")
	cmd.Flags().BoolVar(&fonts, "fonts", false, "build fonts (default: all asset types)")

	return cmd
}

func buildAssets(opts assetBuildOptions) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
	}

	if opts.BuildJS {
		if err := buildJS(cfg.JS, opts.Minify); err != nil {
			return err
		}
	}
	if opts.BuildFonts {
		if err := buildFonts(cfg.Fonts); err != nil {
			return err
		}
	}
	if opts.BuildCSS {
		if err := buildCSSAll(cfg.CSS, opts.Minify); err != nil {
			return err
		}
	}
	return nil
}

// buildFonts downloads self-hosted font files declared in .assets.yaml.
// Each download is skipped when the target file already exists, matching the
// JS download convention. Delete the target to force a re-download.
// Downloaded files land in public_dir/fonts/ and are automatically included
// in the content-hash manifest by publish.Assets on next server start.
func buildFonts(cfg config.FontConfig) error {
	for _, dl := range cfg.Downloads {
		if err := downloadFont(dl); err != nil {
			return err
		}
	}
	return nil
}

func downloadFont(dl config.FontDownload) (err error) {
	if _, err := os.Stat(dl.Target); err == nil {
		fmt.Printf("  ↷ font %s: target exists, skipping (delete to force re-download)\n", dl.Name)
		return nil
	}

	version := resolveVersion(dl.Name, dl.Version)
	url := strings.ReplaceAll(dl.URL, "{version}", version)

	if err := os.MkdirAll(filepath.Dir(dl.Target), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(dl.Target), err)
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer closeOnReturn(&err, resp.Body, "response body for %s", url)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}

	out, err := os.Create(dl.Target)
	if err != nil {
		return fmt.Errorf("create %s: %w", dl.Target, err)
	}
	defer closeOnReturn(&err, out, "file %s", dl.Target)

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write %s: %w", dl.Target, err)
	}
	fmt.Printf("  ✓ downloaded font %s -> %s\n", dl.Name, dl.Target)
	return nil
}

func buildJS(cfg config.JSConfig, minify bool) error {
	if err := removeStaleJSOutputs(cfg); err != nil {
		return err
	}

	for _, dl := range cfg.Downloads {
		if err := downloadJS(dl); err != nil {
			return err
		}
	}

	for _, bundle := range cfg.Bundles {
		if err := bundleJS(bundle, minify); err != nil {
			return err
		}
	}

	return nil
}

// downloadJS fetches a third-party JS file as described by dl.
// The download is skipped when the target file already exists, avoiding
// redundant network round-trips on every dev rebuild. Delete the target
// file to force a re-download on the next asset build.
// The version is read from the {NAME}_VERSION environment variable first,
// falling back to the value in .assets.yaml.
func downloadJS(dl config.JSDownload) (err error) {
	if _, err := os.Stat(dl.Target); err == nil {
		fmt.Printf("  ↷ %s: target exists, skipping (delete to force re-download)\n", dl.Name)
		return nil
	}

	version := resolveVersion(dl.Name, dl.Version)
	url := strings.ReplaceAll(dl.URL, "{version}", version)

	if err := os.MkdirAll(filepath.Dir(dl.Target), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(dl.Target), err)
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer closeOnReturn(&err, resp.Body, "response body for %s", url)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}

	out, err := os.Create(dl.Target)
	if err != nil {
		return fmt.Errorf("create %s: %w", dl.Target, err)
	}
	defer closeOnReturn(&err, out, "file %s", dl.Target)

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write %s: %w", dl.Target, err)
	}
	fmt.Printf("  ✓ downloaded %s@%s -> %s\n", dl.Name, version, dl.Target)
	return nil
}

func bundleJS(cfg config.JSBundle, minify bool) error {
	if cfg.Entry == "" || cfg.Target == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Target), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(cfg.Target), err)
	}

	result := api.Build(api.BuildOptions{
		EntryPoints:       []string{cfg.Entry},
		Bundle:            true,
		Write:             true,
		Outfile:           cfg.Target,
		Platform:          api.PlatformBrowser,
		Format:            api.FormatIIFE,
		TreeShaking:       api.TreeShakingTrue,
		Target:            api.ES2017,
		MinifyWhitespace:  minify,
		MinifyIdentifiers: minify,
		MinifySyntax:      minify,
		LegalComments:     api.LegalCommentsNone,
	})
	if len(result.Errors) > 0 {
		return fmt.Errorf("bundle %s: %s", cfg.Entry, result.Errors[0].Text)
	}
	return nil
}

func removeStaleJSOutputs(cfg config.JSConfig) error {
	managed := make(map[string]bool, len(cfg.Downloads)+len(cfg.Bundles))
	dirs := make(map[string]bool, len(cfg.Downloads)+len(cfg.Bundles))

	for _, dl := range cfg.Downloads {
		if dl.Target == "" {
			continue
		}
		managed[filepath.Clean(dl.Target)] = true
		dirs[filepath.Dir(dl.Target)] = true
	}
	for _, bundle := range cfg.Bundles {
		if bundle.Target == "" {
			continue
		}
		managed[filepath.Clean(bundle.Target)] = true
		dirs[filepath.Dir(bundle.Target)] = true
	}

	for dir := range dirs {
		outputs, err := filepath.Glob(filepath.Join(dir, "*.js"))
		if err != nil {
			return fmt.Errorf("glob %s: %w", dir, err)
		}
		for _, path := range outputs {
			clean := filepath.Clean(path)
			if managed[clean] {
				continue
			}
			if err := os.Remove(clean); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("remove %s: %w", clean, err)
			}
		}
	}

	return nil
}

func buildCSSAll(entries []config.CSSConfig, minify bool) error {
	for _, cfg := range entries {
		if err := buildCSS(cfg, minify); err != nil {
			return err
		}
	}
	return nil
}

func buildCSS(cfg config.CSSConfig, minify bool) error {
	if err := os.MkdirAll(filepath.Dir(cfg.Output), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(cfg.Output), err)
	}
	args := []string{"-i", cfg.Input, "-o", cfg.Output}
	if minify {
		args = append(args, "--minify")
	}
	if err := command(cfg.Tool, args...).Run(); err != nil {
		return fmt.Errorf("%s: %w", cfg.Tool, err)
	}
	return nil
}
