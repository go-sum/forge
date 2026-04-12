package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"
)

type listOptions struct {
	source string
}

func newListCmd() *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all files that would be copied by clone",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&opts.source, "source", "", "forge repository root (default: auto-detect)")

	return cmd
}

// runList prints the files that would be included in a clone operation.
func runList(opts listOptions, w io.Writer) error {
	source := opts.source
	if source == "" {
		var err error
		source, err = findSourceRoot()
		if err != nil {
			return err
		}
	}

	manifestPath := filepath.Join(source, "tools", "starter", "manifest.yaml")
	manifest, err := loadManifest(manifestPath)
	if err != nil {
		return err
	}

	files, err := gitFiles(source)
	if err != nil {
		return err
	}

	count := 0
	for _, rel := range files {
		if isExcluded(rel, manifest.Exclude) {
			continue
		}
		fmt.Fprintln(w, rel)
		count++
	}

	fmt.Fprintf(w, "\nTotal: %d files\n", count)
	return nil
}
