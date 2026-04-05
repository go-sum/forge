package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// packageCheck holds the validation result for a single package.
type packageCheck struct {
	pkg          Package
	localSHA     string
	remoteSHA    string
	version      string // current version from go.mod
	needsRelease bool
}

func newDeployCmd(cfg *config) *cobra.Command {
	var autoFix bool

	cmd := &cobra.Command{
		Use:   "deploy [version]",
		Short: "Validate dependencies and optionally release, tag, and push",
		Long: `Checks that all go-sum/* packages in go.mod are published and up to date.

Without --auto, reports dependency status and exits.
With --auto, releases stale packages, syncs go.prod.mod, bumps APP_VERSION,
commits, tags, and pushes to trigger the CI build.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			explicit := ""
			if len(args) == 1 {
				explicit = args[0]
			}

			// ── 1. Preflight ────────────────────────────────────────
			if err := ensureCleanTree(cfg.repoRoot); err != nil {
				return err
			}
			if err := ensureOnBranch(cfg.repoRoot, "main"); err != nil {
				return err
			}

			currentVersion, nextVersion, err := resolveAppVersion(cfg.repoRoot, explicit)
			if err != nil {
				return err
			}

			gh, err := newGHClient(cfg.owner, cfg.dryRun)
			if err != nil {
				return err
			}

			fmt.Fprintf(logWriter, "Deploy: %s → %s\n\n", currentVersion, nextVersion)

			// ── 2. Dependency Validation (read-only) ────────────────
			pkgs, err := discoverPackages(cfg.repoRoot)
			if err != nil {
				return err
			}

			goModPath := filepath.Join(cfg.repoRoot, "go.mod")
			ctx := context.Background()

			checks, err := validatePackages(ctx, cfg, gh, pkgs, goModPath)
			if err != nil {
				return err
			}

			stale := countStale(checks)

			// Print validation table.
			printCheckTable(cmd, checks)
			fmt.Fprintln(logWriter)

			// ── 3a. Default mode: report and exit ───────────────────
			if !autoFix {
				if stale > 0 {
					fmt.Fprintf(logWriter, "%d package(s) need release. Run with --auto to fix, or release manually.\n", stale)
					return fmt.Errorf("%d stale package(s)", stale)
				}
				fmt.Fprintf(logWriter, "All dependencies valid. Ready to deploy %s.\n", nextVersion)
				fmt.Fprintln(logWriter, "Run with --auto to tag and push.")
				return nil
			}

			// ── 3b. Auto mode: release → sync → commit → tag → push ─
			if stale > 0 {
				fmt.Fprintf(logWriter, "Releasing %d stale package(s)...\n\n", stale)
				for _, chk := range checks {
					if !chk.needsRelease {
						continue
					}
					released, newVer, err := releasePackage(ctx, cfg, gh, chk.pkg, goModPath, "")
					if err != nil {
						return fmt.Errorf("release %s: %w", chk.pkg.Name, err)
					}
					if released {
						fmt.Fprintf(logWriter, "  %s: %s → %s\n", chk.pkg.Name, chk.version, newVer)
					}
				}
				fmt.Fprintln(logWriter)
			}

			// Sync go.prod.mod (always, ensures consistency).
			fmt.Fprintln(logWriter, "Syncing go.prod.mod...")
			if !cfg.dryRun {
				if err := syncProdMod(cfg.repoRoot); err != nil {
					return fmt.Errorf("sync go.prod.mod: %w", err)
				}
			} else {
				fmt.Fprintln(logWriter, "  [dry-run] would regenerate go.prod.mod and run go mod tidy")
			}

			// Update APP_VERSION in .versions.
			if !cfg.dryRun {
				if err := writeDotVersion(cfg.repoRoot, "APP_VERSION", nextVersion); err != nil {
					return fmt.Errorf("update APP_VERSION: %w", err)
				}
				fmt.Fprintf(logWriter, "Updated APP_VERSION to %s\n", nextVersion)
			} else {
				fmt.Fprintf(logWriter, "[dry-run] would update APP_VERSION to %s\n", nextVersion)
			}

			// Commit, tag, and push.
			files := []string{"go.mod", "go.prod.mod", "go.prod.sum", ".versions"}
			commitMsg := fmt.Sprintf("Release %s", nextVersion)

			if !cfg.dryRun {
				if err := gitAdd(cfg.repoRoot, files...); err != nil {
					return fmt.Errorf("git add: %w", err)
				}
				if err := gitCommit(cfg.repoRoot, commitMsg); err != nil {
					return fmt.Errorf("git commit: %w", err)
				}
				if err := gitTag(cfg.repoRoot, nextVersion); err != nil {
					return fmt.Errorf("git tag: %w", err)
				}
				if err := gitPushWithTags(cfg.repoRoot, "origin", "main"); err != nil {
					return fmt.Errorf("git push: %w", err)
				}
				fmt.Fprintf(logWriter, "\nTagged %s — CI build triggered\n", nextVersion)
			} else {
				fmt.Fprintf(logWriter, "\n[dry-run] would commit, tag %s, and push to origin\n", nextVersion)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&autoFix, "auto", false, "auto-release stale packages, sync, tag, and push")

	return cmd
}

// validatePackages checks each package to determine if it needs release.
// This is a read-only operation that does not modify any state.
func validatePackages(ctx context.Context, cfg *config, gh *ghClient, pkgs []Package, goModPath string) ([]packageCheck, error) {
	var checks []packageCheck

	for _, pkg := range pkgs {
		sha, err := splitSubtree(cfg.repoRoot, pkg.Prefix)
		if err != nil {
			return nil, fmt.Errorf("split %s: %w", pkg.Name, err)
		}

		version, err := readGoModVersion(goModPath, pkg.Module)
		if err != nil {
			return nil, fmt.Errorf("read version for %s: %w", pkg.Name, err)
		}

		remoteSHA, err := gh.getRef(ctx, pkg.MirrorRepo, "tags/"+version)
		if err != nil {
			return nil, fmt.Errorf("get remote ref for %s: %w", pkg.Name, err)
		}

		checks = append(checks, packageCheck{
			pkg:          pkg,
			localSHA:     sha,
			remoteSHA:    remoteSHA,
			version:      version,
			needsRelease: sha != remoteSHA,
		})
	}

	return checks, nil
}

func countStale(checks []packageCheck) int {
	n := 0
	for _, c := range checks {
		if c.needsRelease {
			n++
		}
	}
	return n
}

func printCheckTable(cmd *cobra.Command, checks []packageCheck) {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PACKAGE\tVERSION\tLOCAL\tREMOTE\tSTATUS")

	for _, c := range checks {
		status := "ok"
		if c.needsRelease {
			status = "needs release"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			c.pkg.Name, c.version,
			shortSHA(c.localSHA), shortSHA(c.remoteSHA),
			status)
	}
	w.Flush()
}

// ── Git helpers ─────────────────────────────────────────────────────────────

func ensureCleanTree(repoRoot string) error {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if len(strings.TrimSpace(string(out))) > 0 {
		return fmt.Errorf("working tree is not clean:\n%s", string(out))
	}
	return nil
}

func ensureOnBranch(repoRoot, branch string) error {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git rev-parse: %w", err)
	}
	current := strings.TrimSpace(string(out))
	if current != branch {
		return fmt.Errorf("must be on branch %q, currently on %q", branch, current)
	}
	return nil
}

func gitAdd(repoRoot string, files ...string) error {
	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add %s: %w", strings.Join(files, " "), err)
	}
	return nil
}

func gitCommit(repoRoot, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

func gitTag(repoRoot, version string) error {
	cmd := exec.Command("git", "tag", version)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git tag %s: %w", version, err)
	}
	return nil
}

func gitPushWithTags(repoRoot, remote, branch string) error {
	cmd := exec.Command("git", "push", remote, branch, "--tags")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push %s %s --tags: %w", remote, branch, err)
	}
	return nil
}
