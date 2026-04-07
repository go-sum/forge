package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// releasePackage releases a single package by subtree-splitting and pushing to its mirror repo.
// If explicitVersion is empty, the patch version is auto-incremented.
// Returns whether a release was made and the new version string.
func releasePackage(ctx context.Context, cfg *config, gh *ghClient, pkg Package, goModPath, explicitVersion string) (bool, string, error) {
	if err := ensureRepoExists(ctx, gh, pkg); err != nil {
		return false, "", err
	}

	sha, err := splitSubtree(cfg.repoRoot, pkg.Prefix)
	if err != nil {
		return false, "", err
	}
	fmt.Fprintf(logWriter, "  split SHA: %s\n", sha)

	// When no explicit version is given, skip if nothing changed since the last release.
	if explicitVersion == "" {
		currentVersion, err := readGoModVersion(goModPath, pkg.Module)
		if err != nil {
			return false, "", err
		}
		remoteTagSHA, err := gh.getRef(ctx, pkg.MirrorRepo, "tags/"+currentVersion)
		if err != nil {
			return false, "", err
		}
		if remoteTagSHA == sha {
			fmt.Fprintf(logWriter, "No changes since %s, skipping %s\n", currentVersion, pkg.Name)
			return false, currentVersion, nil
		}
	}

	version, err := resolveReleaseVersion(goModPath, pkg.Module, explicitVersion)
	if err != nil {
		return false, "", err
	}

	fmt.Fprintf(logWriter, "Releasing %s to %s/%s as %s\n", pkg.Prefix, cfg.owner, pkg.MirrorRepo, version)

	refs := []string{"refs/heads/main", "refs/tags/" + version}
	if err := pushGit(cfg.repoRoot, gh.token, cfg.owner, pkg.MirrorRepo, sha, refs, cfg.dryRun); err != nil {
		return false, "", err
	}

	if err := gh.createRelease(ctx, pkg.MirrorRepo, version, pkg.Prefix); err != nil {
		return false, "", err
	}

	// Update go.mod with the released version.
	if !cfg.dryRun {
		if err := writeGoModVersion(goModPath, pkg.Module, version); err != nil {
			return false, "", fmt.Errorf("update go.mod: %w", err)
		}
		fmt.Fprintf(logWriter, "  updated %s to %s in go.mod\n", pkg.Module, version)
	} else {
		fmt.Fprintf(logWriter, "  [dry-run] would update %s to %s in go.mod\n", pkg.Module, version)
	}

	return true, version, nil
}

func newReleaseCmd(cfg *config) *cobra.Command {
	return &cobra.Command{
		Use:   "release <name|all> [version]",
		Short: "Release a versioned package to its mirror repo",
		Long: `Release a package by subtree-splitting and pushing to its mirror repo.

If version is omitted, the patch version from go.mod is auto-incremented.
If version is specified, it must be greater than the current version in go.mod.
After a successful release, go.mod is updated with the new version.

Use 'all' to release every discovered package; explicit version is not allowed with 'all'.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			nameOrAll := args[0]
			explicit := ""
			if len(args) == 2 {
				if nameOrAll == "all" {
					return fmt.Errorf("explicit version cannot be used with 'all'")
				}
				explicit = args[1]
			}

			gh, err := newGHClient(cfg.owner, cfg.dryRun)
			if err != nil {
				return err
			}

			pkgs, err := resolvePackages(cfg.repoRoot, nameOrAll)
			if err != nil {
				return err
			}

			goModPath := filepath.Join(cfg.repoRoot, "go.mod")
			ctx := context.Background()

			for _, pkg := range pkgs {
				if _, _, err = releasePackage(ctx, cfg, gh, pkg, goModPath, explicit); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
