package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newPushCmd(cfg *config) *cobra.Command {
	return &cobra.Command{
		Use:   "push <name|all>",
		Short: "Subtree split and push a package to its mirror repo",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gh, err := newGHClient(cfg.owner, cfg.dryRun)
			if err != nil {
				return err
			}

			pkgs, err := resolvePackages(cfg.repoRoot, args[0])
			if err != nil {
				return err
			}

			ctx := context.Background()

			for _, pkg := range pkgs {
				if err := ensureRepoExists(ctx, gh, pkg); err != nil {
					return err
				}

				fmt.Fprintf(logWriter, "Syncing %s to %s/%s@main\n", pkg.Prefix, cfg.owner, pkg.MirrorRepo)

				sha, err := splitSubtree(cfg.repoRoot, pkg.Prefix)
				if err != nil {
					return err
				}
				fmt.Fprintf(logWriter, "  split SHA: %s\n", sha)

				remoteSHA, err := gh.getRef(ctx, pkg.MirrorRepo, "heads/main")
				if err != nil {
					return err
				}
				if remoteSHA == sha {
					fmt.Fprintf(logWriter, "  already in sync, skipping\n")
					continue
				}

				if err := pushGit(cfg.repoRoot, gh.token, cfg.owner, pkg.MirrorRepo, sha,
					[]string{"refs/heads/main"}, cfg.dryRun); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
