package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// dsn returns the DATABASE_URL environment variable.
// Returns an error if DATABASE_URL is not set.
func dsn() (string, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return "", fmt.Errorf("DATABASE_URL environment variable is required")
	}
	return url, nil
}

func main() {
	root := &cobra.Command{
		Use:           "db",
		Short:         "Database migration management",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newMigrateCmd(),
		newRollbackCmd(),
		newStatusCmd(),
		newCreateCmd(),
		newComposeCmd(),
	)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
