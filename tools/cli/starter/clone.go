package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type cloneOptions struct {
	source string
	target string
	module string
}

func newCloneCmd() *cobra.Command {
	var opts cloneOptions

	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone the forge starter template into a new application directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClone(opts, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&opts.target, "target", "", "destination directory path (required)")
	cmd.Flags().StringVar(&opts.module, "module", "", "new Go module path, e.g. github.com/myorg/myapp (required)")
	cmd.Flags().StringVar(&opts.source, "source", "", "forge repository root (default: auto-detect)")

	_ = cmd.MarkFlagRequired("target")
	_ = cmd.MarkFlagRequired("module")

	return cmd
}

// runClone performs the full clone operation and writes a summary to w.
func runClone(opts cloneOptions, w io.Writer) error {
	source := opts.source
	if source == "" {
		var err error
		source, err = findSourceRoot()
		if err != nil {
			return err
		}
	}

	target := filepath.Clean(opts.target)

	// Reject target if it already exists and is non-empty.
	if entries, err := os.ReadDir(target); err == nil && len(entries) > 0 {
		return fmt.Errorf("clone: target directory %q already exists and is non-empty; remove it first", target)
	}

	manifestPath := filepath.Join(source, "tools", "starter", "manifest.yaml")
	manifest, err := loadManifest(manifestPath)
	if err != nil {
		return err
	}

	// Step 1: walk and copy files.
	filesCopied, err := copyFiles(source, target, manifest.Exclude)
	if err != nil {
		return err
	}

	// Step 2: apply renames.
	filesRenamed, err := applyRenames(target, manifest.Rename)
	if err != nil {
		return err
	}

	// Step 3: strip monorepo-only content blocks from Markdown files.
	filesStripped, err := stripMonorepoBlocks(target)
	if err != nil {
		return err
	}

	// Step 4: rewrite module paths.
	filesRewritten, err := rewriteModule(target, manifest.ModuleRewrite.From, opts.module)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Clone complete.\n")
	fmt.Fprintf(w, "  Files copied:    %d\n", filesCopied)
	fmt.Fprintf(w, "  Files renamed:   %d\n", filesRenamed)
	fmt.Fprintf(w, "  Blocks stripped: %d\n", filesStripped)
	fmt.Fprintf(w, "  Files rewritten: %d\n", filesRewritten)
	fmt.Fprintf(w, "\nNext steps:\n")
	fmt.Fprintf(w, "  cd %s\n", target)
	fmt.Fprintf(w, "  go mod tidy\n")
	fmt.Fprintf(w, "  make db-migrate\n")
	fmt.Fprintf(w, "  make dev\n")
	return nil
}

// copyFiles uses git to enumerate source files (respecting .gitignore), then
// copies files not matched by the manifest exclude list to target.
// Returns the number of files copied.
func copyFiles(source, target string, excludes []string) (int, error) {
	files, err := gitFiles(source)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, rel := range files {
		if isExcluded(rel, excludes) {
			continue
		}
		if err := copyFile(filepath.Join(source, rel), filepath.Join(target, rel)); err != nil {
			return count, fmt.Errorf("copy %s: %w", rel, err)
		}
		count++
	}
	return count, nil
}

// gitFiles returns the list of file paths (relative to dir) that git would
// include: all tracked files plus untracked files not covered by .gitignore.
// Files that are tracked in the index but deleted on disk (unstaged deletions)
// are excluded — they would not be delivered by a fresh git clone.
func gitFiles(dir string) ([]string, error) {
	cmd := exec.Command("git", "-C", dir, "ls-files",
		"--cached",
		"--others",
		"--exclude-standard",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files in %s: %w", dir, err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if line == "" {
			continue
		}
		rel := filepath.ToSlash(line)
		// Skip files tracked in the index but not present on disk.
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			continue
		}
		files = append(files, rel)
	}
	return files, nil
}

// copyFile copies a single file from src to dst, creating parent directories.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// applyRenames renames files in the target directory per the rename rules.
// Returns the number of files renamed.
func applyRenames(target string, rules []RenameRule) (int, error) {
	count := 0
	for _, rule := range rules {
		src := filepath.Join(target, filepath.FromSlash(rule.From))
		dst := filepath.Join(target, filepath.FromSlash(rule.To))
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		if err := os.Rename(src, dst); err != nil {
			return count, fmt.Errorf("rename %s -> %s: %w", rule.From, rule.To, err)
		}
		count++
	}
	return count, nil
}

// rewriteModule replaces fromModule with toModule in go.mod and all *.go files.
// Returns the number of files that were modified.
func rewriteModule(target, fromModule, toModule string) (int, error) {
	count := 0

	// Rewrite go.mod: replace the module directive line.
	goModPath := filepath.Join(target, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		modified, err := rewriteFileStrings(goModPath, map[string]string{
			"module " + fromModule: "module " + toModule,
			// Also rewrite any replacement/require lines referencing the old module.
			fromModule + "/": toModule + "/",
		})
		if err != nil {
			return count, fmt.Errorf("rewrite go.mod: %w", err)
		}
		if modified {
			count++
		}
	}

	// Rewrite all *.go files: replace import paths.
	err := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		modified, err := rewriteFileStrings(path, map[string]string{
			`"` + fromModule + `/`: `"` + toModule + `/`,
		})
		if err != nil {
			return fmt.Errorf("rewrite %s: %w", path, err)
		}
		if modified {
			count++
		}
		return nil
	})
	return count, err
}

// stripMonorepoBlocks removes content between <!-- monorepo-only-start --> and
// <!-- monorepo-only-end --> markers from all Markdown files in target.
// Returns the number of files modified.
func stripMonorepoBlocks(target string) (int, error) {
	const startMarker = "<!-- monorepo-only-start -->"
	const endMarker = "<!-- monorepo-only-end -->"

	count := 0
	err := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		if !strings.Contains(content, startMarker) {
			return nil
		}
		// Strip every block between start and end markers (inclusive).
		result := content
		for {
			start := strings.Index(result, startMarker)
			if start == -1 {
				break
			}
			end := strings.Index(result[start:], endMarker)
			if end == -1 {
				break
			}
			end += start + len(endMarker)
			// Also consume a trailing newline if present.
			if end < len(result) && result[end] == '\n' {
				end++
			}
			result = result[:start] + result[end:]
		}
		if result == content {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(result), info.Mode()); err != nil {
			return fmt.Errorf("strip %s: %w", path, err)
		}
		count++
		return nil
	})
	return count, err
}

// rewriteFileStrings performs string replacements in a file.
// Returns true if the file was modified.
func rewriteFileStrings(path string, replacements map[string]string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	content := string(data)
	original := content
	for old, new := range replacements {
		content = strings.ReplaceAll(content, old, new)
	}

	if content == original {
		return false, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if err := os.WriteFile(path, []byte(content), info.Mode()); err != nil {
		return false, err
	}
	return true, nil
}
