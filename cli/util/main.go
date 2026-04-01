package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "util",
		Short:         "Development utilities",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newHashAirCSPCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
