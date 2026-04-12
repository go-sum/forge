package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:          "starter",
		Short:        "Bootstrap a new app from the forge starter template",
		SilenceUsage: true,
	}

	root.AddCommand(
		newCloneCmd(),
		newVerifyCmd(),
		newListCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// findSourceRoot locates the forge repository root by checking, in order:
//  1. The FORGE_ROOT environment variable.
//  2. Walking up from the executable's directory.
//  3. The current working directory.
func findSourceRoot() (string, error) {
	if v := os.Getenv("FORGE_ROOT"); v != "" {
		if isForgeRoot(v) {
			return v, nil
		}
		return "", fmt.Errorf("FORGE_ROOT=%q does not look like a forge root (missing go.work or tools/starter/manifest.yaml)", v)
	}

	exe, err := os.Executable()
	if err == nil {
		if root, ok := walkUpForRoot(filepath.Dir(exe)); ok {
			return root, nil
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("findSourceRoot: getwd: %w", err)
	}
	if isForgeRoot(cwd) {
		return cwd, nil
	}
	if root, ok := walkUpForRoot(cwd); ok {
		return root, nil
	}

	return "", fmt.Errorf("findSourceRoot: cannot locate forge root; set FORGE_ROOT or run from within the forge repository")
}

// walkUpForRoot walks up the directory tree from dir looking for a forge root.
func walkUpForRoot(dir string) (string, bool) {
	for {
		if isForgeRoot(dir) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// isForgeRoot returns true if dir contains both go.work and
// tools/starter/manifest.yaml.
func isForgeRoot(dir string) bool {
	goWork := filepath.Join(dir, "go.work")
	manifest := filepath.Join(dir, "tools", "starter", "manifest.yaml")
	return fileExists(goWork) && fileExists(manifest)
}

// fileExists returns true if the path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// isExcluded reports whether relPath should be excluded according to the
// manifest's exclude list. Directory entries are matched by prefix.
func isExcluded(relPath string, excludes []string) bool {
	// Always skip .git.
	if relPath == ".git" || strings.HasPrefix(relPath, ".git/") {
		return true
	}
	for _, ex := range excludes {
		if ex == relPath {
			return true
		}
		// Directory pattern: ends with "/" — match everything inside.
		if strings.HasSuffix(ex, "/") && strings.HasPrefix(relPath, ex) {
			return true
		}
		// Also allow patterns without trailing slash to match as directory prefix.
		if !strings.Contains(ex, ".") && (relPath == ex || strings.HasPrefix(relPath, ex+"/")) {
			return true
		}
	}
	return false
}
