package main

import (
	"fmt"

	"github.com/go-sum/server/database/migrate"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new empty migration file",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreate,
	}
	cmd.Flags().String("dir", "db/migrations", "migrations directory")
	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	name := args[0]

	path, err := migrate.Create(dir, name)
	if err != nil {
		return err
	}
	fmt.Printf("Created: %s\n", path)
	return nil
}
