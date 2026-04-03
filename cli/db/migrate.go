package main

import (
	"context"
	"fmt"

	"github.com/go-sum/server/database/migrate"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply all pending migrations",
		Args:  cobra.NoArgs,
		RunE:  runMigrate,
	}
	cmd.Flags().String("dir", "db/migrations", "migrations directory")
	return cmd
}

func runMigrate(cmd *cobra.Command, _ []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	connStr, err := dsn()
	if err != nil {
		return err
	}

	ctx := context.Background()
	fmt.Println("Applying pending migrations...")
	if err := migrate.Up(ctx, connStr, dir); err != nil {
		return err
	}
	fmt.Println("Migrations applied successfully.")
	return nil
}
