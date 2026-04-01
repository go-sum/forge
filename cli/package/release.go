package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newReleaseCmd(cfg *config) *cobra.Command {
	return &cobra.Command{
		Use:   "release <name> [version]",
		Short: "Release a versioned package to its mirror repo",
		Long: `Release a package by subtree-splitting and pushing to its mirror repo.

If version is omitted, the patch version from go.mod is auto-incremented.
If version is specified, it must be greater than the current version in go.mod.
After a successful release, go.mod is updated with the new version.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			explicit := ""
			if len(args) == 2 {
				explicit = args[1]
			}

			gh, err := newGHClient(cfg.owner, cfg.dryRun)
			if err != nil {
				return err
			}

			pkg, err := discoverPackage(cfg.repoRoot, name)
			if err != nil {
				return err
			}

			goModPath := filepath.Join(cfg.repoRoot, "go.mod")
			ctx := context.Background()

			if err := ensureRepoExists(ctx, gh, pkg); err != nil {
				return err
			}

			sha, err := splitSubtree(cfg.repoRoot, pkg.Prefix)
			if err != nil {
				return err
			}
			fmt.Fprintf(logWriter, "  split SHA: %s\n", sha)

			// When no explicit version is given, skip if nothing has been pushed since the last release.
			if explicit == "" {
				currentVersion, err := readGoModVersion(goModPath, pkg.Module)
				if err != nil {
					return err
				}
				remoteTagSHA, err := gh.getRef(ctx, pkg.MirrorRepo, "tags/"+currentVersion)
				if err != nil {
					return err
				}
				if remoteTagSHA == sha {
					fmt.Fprintf(logWriter, "No changes since %s, skipping %s\n", currentVersion, pkg.Name)
					return nil
				}
			}

			version, err := resolveReleaseVersion(goModPath, pkg.Module, explicit)
			if err != nil {
				return err
			}

			fmt.Fprintf(logWriter, "Releasing %s to %s/%s as %s\n", pkg.Prefix, cfg.owner, pkg.MirrorRepo, version)

			refs := []string{"refs/heads/main", "refs/tags/" + version}
			if err := pushGit(cfg.repoRoot, gh.token, cfg.owner, pkg.MirrorRepo, sha, refs, cfg.dryRun); err != nil {
				return err
			}

			if err := gh.createRelease(ctx, pkg.MirrorRepo, version, pkg.Prefix); err != nil {
				return err
			}

			// Update go.mod with the released version.
			if !cfg.dryRun {
				if err := writeGoModVersion(goModPath, pkg.Module, version); err != nil {
					return fmt.Errorf("update go.mod: %w", err)
				}
				fmt.Fprintf(logWriter, "  updated %s to %s in go.mod\n", pkg.Module, version)
			} else {
				fmt.Fprintf(logWriter, "  [dry-run] would update %s to %s in go.mod\n", pkg.Module, version)
			}

			return nil
		},
	}
}
