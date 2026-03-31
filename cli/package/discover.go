package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Package represents a discovered package under pkg/.
type Package struct {
	Name       string // directory name: "auth"
	Dir        string // absolute path: /src/forge/pkg/auth
	Module     string // module path: "github.com/go-sum/auth"
	Prefix     string // subtree prefix: "pkg/auth"
	MirrorRepo string // mirror repo name: "auth"
}

// discoverPackages scans pkg/*/go.mod and returns all discovered packages.
func discoverPackages(repoRoot string) ([]Package, error) {
	pattern := filepath.Join(repoRoot, "pkg", "*", "go.mod")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob %s: %w", pattern, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no packages found under %s/pkg/", repoRoot)
	}

	var pkgs []Package
	for _, gomod := range matches {
		dir := filepath.Dir(gomod)
		name := filepath.Base(dir)

		mod, err := parseModulePath(gomod)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", gomod, err)
		}

		// Mirror repo name is the last segment of the module path.
		parts := strings.Split(mod, "/")
		mirror := parts[len(parts)-1]

		pkgs = append(pkgs, Package{
			Name:       name,
			Dir:        dir,
			Module:     mod,
			Prefix:     filepath.Join("pkg", name),
			MirrorRepo: mirror,
		})
	}

	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Name < pkgs[j].Name
	})

	return pkgs, nil
}

// discoverPackage returns a single package by name, or an error if not found.
func discoverPackage(repoRoot, name string) (Package, error) {
	pkgs, err := discoverPackages(repoRoot)
	if err != nil {
		return Package{}, err
	}

	for _, p := range pkgs {
		if p.Name == name {
			return p, nil
		}
	}

	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.Name
	}
	return Package{}, fmt.Errorf("unknown package %q (available: %s)", name, strings.Join(names, ", "))
}

// resolvePackages resolves "all" to all packages, or a single name to one package.
func resolvePackages(repoRoot, nameOrAll string) ([]Package, error) {
	if nameOrAll == "all" {
		return discoverPackages(repoRoot)
	}
	p, err := discoverPackage(repoRoot, nameOrAll)
	if err != nil {
		return nil, err
	}
	return []Package{p}, nil
}

// parseModulePath reads the module directive from a go.mod file.
func parseModulePath(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no module directive found in %s", goModPath)
}
