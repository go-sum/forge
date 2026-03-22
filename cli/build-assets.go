package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"github.com/go-sum/componentry/assetconfig"
	"strings"
	"syscall"

	"github.com/evanw/esbuild/pkg/api"
)

type assetBuildOptions struct {
	ConfigPath   string
	Minify       bool
	BuildCSS     bool
	BuildJS      bool
	BuildSprites bool
}

func runBuildAssets() {
	fs := flag.NewFlagSet("build-assets", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := fs.String("config", assetconfig.DefaultConfigPath, "path to assets config file")
	minify := fs.Bool("minify", false, "minify compiled CSS and JS")
	cssOnly := fs.Bool("css-only", false, "build only CSS")
	jsOnly := fs.Bool("js-only", false, "build only JS assets")
	spritesOnly := fs.Bool("sprites-only", false, "build only SVG sprites")
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	opts := assetBuildOptions{
		ConfigPath:   *configPath,
		Minify:       *minify,
		BuildCSS:     true,
		BuildJS:      true,
		BuildSprites: true,
	}

	selected := 0
	for _, only := range []bool{*cssOnly, *jsOnly, *spritesOnly} {
		if only {
			selected++
		}
	}
	if selected > 1 {
		fmt.Fprintln(os.Stderr, "error: choose at most one of --css-only, --js-only, --sprites-only")
		os.Exit(1)
	}
	if selected == 1 {
		opts.BuildCSS = *cssOnly
		opts.BuildJS = *jsOnly
		opts.BuildSprites = *spritesOnly
	}

	if err := buildAssets(opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// runDev delegates all watching, rebuilding, and hot-reload to air.
// air's pre_cmd in .air.toml runs build-assets before each server rebuild,
// so CSS, JS, sprite, and config changes all trigger a consistent pipeline
// and a server restart with freshly hashed assets.
func runDev() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	airCmd := commandContext(ctx, "air", "-c", ".air.toml")
	if err := airCmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "error starting air:", err)
		os.Exit(1)
	}

	if err := waitNamed("air", airCmd); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func buildAssets(opts assetBuildOptions) error {
	cfg, err := assetconfig.Load(opts.ConfigPath)
	if err != nil {
		return err
	}

	if opts.BuildJS {
		if err := buildJS(cfg.JS, opts.Minify); err != nil {
			return err
		}
	}
	if opts.BuildSprites {
		if err := buildSVGSprites([]string{"--config", opts.ConfigPath}); err != nil {
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

func buildJS(cfg assetconfig.JSConfig, minify bool) error {
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
func downloadJS(dl assetconfig.JSDownload) error {
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}

	out, err := os.Create(dl.Target)
	if err != nil {
		return fmt.Errorf("create %s: %w", dl.Target, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write %s: %w", dl.Target, err)
	}
	fmt.Printf("  ✓ downloaded %s@%s -> %s\n", dl.Name, version, dl.Target)
	return nil
}

// resolveVersion returns the version from the {NAME}_VERSION environment
// variable (e.g. HTMX_VERSION), falling back to the value in .assets.yaml.
func resolveVersion(name, defaultVersion string) string {
	if v := strings.TrimSpace(os.Getenv(strings.ToUpper(name) + "_VERSION")); v != "" {
		return v
	}
	return defaultVersion
}

func bundleJS(cfg assetconfig.JSBundle, minify bool) error {
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

func removeStaleJSOutputs(cfg assetconfig.JSConfig) error {
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

func buildCSSAll(entries []assetconfig.CSSConfig, minify bool) error {
	for _, cfg := range entries {
		if err := buildCSS(cfg, minify); err != nil {
			return err
		}
	}
	return nil
}

func buildCSS(cfg assetconfig.CSSConfig, minify bool) error {
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s -> %s: %w", src, dst, err)
	}
	return nil
}

func command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func commandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func waitNamed(name string, cmd *exec.Cmd) error {
	if err := cmd.Wait(); err != nil {
		if isExpectedExit(err) {
			return nil
		}
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func isExpectedExit(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		return false
	}
	return status.Signaled() && (status.Signal() == syscall.SIGTERM || status.Signal() == syscall.SIGKILL)
}
