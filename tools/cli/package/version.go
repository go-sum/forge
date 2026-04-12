package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// semver represents a parsed semantic version.
type semver struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
}

var semverParseRe = regexp.MustCompile(`^v([0-9]+)\.([0-9]+)\.([0-9]+)([.-].+)?$`)

// parseSemver parses a version string like "v1.2.3" or "v1.2.3-rc.1".
func parseSemver(s string) (semver, error) {
	m := semverParseRe.FindStringSubmatch(s)
	if m == nil {
		return semver{}, fmt.Errorf("invalid semver: %q", s)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return semver{Major: major, Minor: minor, Patch: patch, Prerelease: m[4]}, nil
}

// String returns the version as "vMAJOR.MINOR.PATCH[prerelease]".
func (v semver) String() string {
	return fmt.Sprintf("v%d.%d.%d%s", v.Major, v.Minor, v.Patch, v.Prerelease)
}

// bumpPatch returns a new version with the patch number incremented.
func (v semver) bumpPatch() semver {
	return semver{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
}

// greaterThan returns true if v is strictly greater than other.
func (v semver) greaterThan(other semver) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

// readGoModVersion reads the version of a module from go.mod's require block.
func readGoModVersion(goModPath, modulePath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	inRequire := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "require (" {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}
		if !inRequire {
			continue
		}

		// Require lines: "github.com/go-sum/auth v0.0.0"
		// Skip replace lines which contain "=>"
		if strings.Contains(line, "=>") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[0] == modulePath {
			return parts[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module %s not found in %s", modulePath, goModPath)
}

// writeGoModVersion updates the version of a module in go.mod's require block.
func writeGoModVersion(goModPath, modulePath, newVersion string) error {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "=>") {
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) >= 2 && parts[0] == modulePath {
			// Preserve leading whitespace.
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			lines[i] = indent + modulePath + " " + newVersion
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("module %s not found in %s", modulePath, goModPath)
	}

	return os.WriteFile(goModPath, []byte(strings.Join(lines, "\n")), 0644)
}

// readDotVersion reads a single key from the .versions file (KEY=VALUE format).
func readDotVersion(repoRoot, key string) (string, error) {
	path := filepath.Join(repoRoot, ".versions")
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(k) == key {
			return strings.TrimSpace(v), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("key %s not found in %s", key, path)
}

// writeDotVersion updates a single key in the .versions file, preserving other entries.
func writeDotVersion(repoRoot, key, value string) error {
	path := filepath.Join(repoRoot, ".versions")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		k, _, ok := strings.Cut(trimmed, "=")
		if ok && strings.TrimSpace(k) == key {
			lines[i] = key + "=" + value
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("key %s not found in %s", key, path)
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// resolveAppVersion reads APP_VERSION from .versions and resolves the next version.
// If explicit is non-empty, validates it is > current. Otherwise, bumps patch.
// Returns the current version string and the next version string.
func resolveAppVersion(repoRoot, explicit string) (string, string, error) {
	currentStr, err := readDotVersion(repoRoot, "APP_VERSION")
	if err != nil {
		return "", "", err
	}

	current, err := parseSemver(currentStr)
	if err != nil {
		return "", "", fmt.Errorf("APP_VERSION in .versions: %w", err)
	}

	if explicit == "" {
		next := current.bumpPatch()
		return currentStr, next.String(), nil
	}

	next, err := parseSemver(explicit)
	if err != nil {
		return "", "", err
	}

	if !next.greaterThan(current) {
		return "", "", fmt.Errorf("version %s must be greater than current %s", next, current)
	}

	return currentStr, next.String(), nil
}

// resolveReleaseVersion determines the version to release.
// If explicit is non-empty, validates it is > current. Otherwise, bumps patch.
func resolveReleaseVersion(goModPath, modulePath, explicit string) (string, error) {
	currentStr, err := readGoModVersion(goModPath, modulePath)
	if err != nil {
		return "", err
	}

	current, err := parseSemver(currentStr)
	if err != nil {
		return "", fmt.Errorf("current version in go.mod: %w", err)
	}

	if explicit == "" {
		next := current.bumpPatch()
		return next.String(), nil
	}

	next, err := parseSemver(explicit)
	if err != nil {
		return "", err
	}

	if !next.greaterThan(current) {
		return "", fmt.Errorf("version %s must be greater than current %s in go.mod", next, current)
	}

	return next.String(), nil
}
