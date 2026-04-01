package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "create",
		Short:         "Out-of-band record creation utilities",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newAdminCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
