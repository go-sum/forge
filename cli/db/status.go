package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/go-sum/server/database/migrate"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Args:  cobra.NoArgs,
		RunE:  runStatus,
	}
	cmd.Flags().String("dir", "db/migrations", "migrations directory")
	return cmd
}

func runStatus(cmd *cobra.Command, _ []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	connStr, err := dsn()
	if err != nil {
		return err
	}

	ctx := context.Background()
	statuses, err := migrate.Status(ctx, connStr, dir)
	if err != nil {
		return err
	}

	if len(statuses) == 0 {
		fmt.Println("No migrations found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VERSION\tSTATUS\tSOURCE")
	for _, s := range statuses {
		status := "pending"
		if s.Applied {
			status = "applied"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\n", s.Version, status, s.Source)
	}
	return w.Flush()
}
