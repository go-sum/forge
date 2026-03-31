package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "build",
		Short:         "Asset build pipeline",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newAssetsCmd(), newDocsCmd(), newSpritesCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
