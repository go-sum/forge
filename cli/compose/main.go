package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "compose",
		Short:         "Docker Compose service lifecycle utilities",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newRunCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
