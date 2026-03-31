package main

import (
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newListCmd(cfg *config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all discovered packages under pkg/",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pkgs, err := discoverPackages(cfg.repoRoot)
			if err != nil {
				return err
			}

			goModPath := filepath.Join(cfg.repoRoot, "go.mod")

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tVERSION\tMODULE\tPREFIX")
			for _, p := range pkgs {
				ver, _ := readGoModVersion(goModPath, p.Module)
				if ver == "" {
					ver = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, ver, p.Module, p.Prefix)
			}
			return w.Flush()
		},
	}
}
