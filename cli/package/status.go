package main

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newStatusCmd(cfg *config) *cobra.Command {
	return &cobra.Command{
		Use:   "status <name|all>",
		Short: "Compare local split SHA with the remote mirror",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gh, err := newGHClient(cfg.owner, cfg.dryRun)
			if err != nil {
				return err
			}
			if gh == nil {
				return fmt.Errorf("GH_TOKEN is required for the status command")
			}

			pkgs, err := resolvePackages(cfg.repoRoot, args[0])
			if err != nil {
				return err
			}

			ctx := context.Background()
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tLOCAL\tREMOTE\tSTATUS")

			for _, pkg := range pkgs {
				localSHA, err := splitSubtree(cfg.repoRoot, pkg.Prefix)
				if err != nil {
					return err
				}

				remoteSHA, err := gh.getRef(ctx, pkg.MirrorRepo, "heads/main")
				if err != nil {
					return err
				}

				status := "in sync"
				if remoteSHA == "" {
					status = "not pushed"
				} else if localSHA != remoteSHA {
					status = "out of sync"
				}

				local := shortSHA(localSHA)
				remote := shortSHA(remoteSHA)
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", pkg.Name, local, remote, status)
			}

			return w.Flush()
		},
	}
}

func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
