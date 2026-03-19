// Package assets provides content-hash-based cache-busting for public files.
//
// New walks the built public directory, computes an 8-char SHA-256 hash of each
// file's contents, and builds a URL manifest. Path then returns versioned
// URLs like "/public/css/app.css?v=abc12345".
//
// Hashing runs in both development and production so that cache headers,
// CSP rules, and asset URL behaviour are identical across environments.
// Because air restarts the server after each asset rebuild, the manifest is
// always recomputed against the latest files.
//
// If publicDir does not exist (e.g. assets not yet built), New returns an
// empty manifest rather than an error; Path falls back to bare URLs.
//
// # Usage
//
// The package provides both an instance API for test isolation and a
// package-level default for ergonomic use in views:
//
//	// main.go — called once at startup
//	assets.MustInit("public", "/public")
//
//	// views — no instance required
//	src := assets.Path("css/app.css")
package assets

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Assets values are immutable after construction; concurrent reads are safe.
type Assets struct {
	manifest  map[string]string
	urlPrefix string
}

// New builds an Assets manifest by walking publicDir and hashing each file.
// publicDir is the filesystem path to the built public files (e.g. "public"),
// prefix is the URL prefix (e.g. "/public").
//
// If publicDir does not exist, New returns an empty manifest without error.
func New(publicDir, prefix string) (*Assets, error) {
	a := &Assets{
		manifest:  make(map[string]string),
		urlPrefix: prefix,
	}

	err := filepath.WalkDir(publicDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		hash, err := hashFile(path)
		if err != nil {
			return fmt.Errorf("assets: hashing %s: %w", path, err)
		}

		rel, err := filepath.Rel(publicDir, path)
		if err != nil {
			return fmt.Errorf("assets: rel path for %s: %w", path, err)
		}
		rel = filepath.ToSlash(rel)

		a.manifest[rel] = prefix + "/" + rel + "?v=" + hash
		return nil
	})
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return a, nil
		}
		return nil, err
	}

	return a, nil
}

// Must panics if err is non-nil. Intended for wrapping New in main().
func Must(a *Assets, err error) *Assets {
	if err != nil {
		panic(fmt.Sprintf("assets: %v", err))
	}
	return a
}

// Path returns the versioned URL for the given relative asset name.
// For example, Path("css/app.css") returns "/public/css/app.css?v=abc12345".
//
// If the key is not found in the manifest, Path returns a bare fallback URL.
func (a *Assets) Path(name string) string {
	if url, ok := a.manifest[name]; ok {
		return url
	}
	prefix := a.urlPrefix
	if prefix == "" {
		prefix = "/public"
	}
	return prefix + "/" + strings.TrimPrefix(name, "/")
}

// --- Package-level convenience API (backed by Default) ---

// Default is the package-level Assets instance. It is set by MustInit or Init.
// Views should call the package-level Path function rather than accessing Default directly.
var Default = &Assets{}

// Init builds the package-level Default manifest.
func Init(publicDir, prefix string) error {
	a, err := New(publicDir, prefix)
	if err != nil {
		return err
	}
	Default = a
	return nil
}

// MustInit calls Init and panics on error. Intended for use in main().
func MustInit(publicDir, prefix string) {
	if err := Init(publicDir, prefix); err != nil {
		panic(fmt.Sprintf("assets.MustInit: %v", err))
	}
}

// Path delegates to Default.Path for ergonomic use in views.
func Path(name string) string { return Default.Path(name) }

// hashFile returns the first 8 hex characters of the SHA-256 of the file at path.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:8], nil
}
