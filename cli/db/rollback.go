package main

import (
	"context"
	"fmt"

	"github.com/go-sum/server/database/migrate"
	"github.com/spf13/cobra"
)

func newRollbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback the last applied migration",
		Args:  cobra.NoArgs,
		RunE:  runRollback,
	}
	cmd.Flags().String("dir", "db/migrations", "migrations directory")
	return cmd
}

func runRollback(cmd *cobra.Command, _ []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	connStr, err := dsn()
	if err != nil {
		return err
	}

	ctx := context.Background()
	fmt.Println("Rolling back last migration...")
	if err := migrate.Down(ctx, connStr, dir); err != nil {
		return err
	}
	fmt.Println("Rollback complete.")
	return nil
}
