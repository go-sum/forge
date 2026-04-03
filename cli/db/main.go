package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// dsn builds a PostgreSQL connection string from standard PG* environment
// variables. Returns an error if PGHOST is not set.
func dsn() (string, error) {
	host := os.Getenv("PGHOST")
	if host == "" {
		return "", fmt.Errorf("PGHOST environment variable is required")
	}
	port := os.Getenv("PGPORT")
	if port == "" {
		port = "5432"
	}
	name := os.Getenv("PGDATABASE")
	if name == "" {
		name = "postgres"
	}
	user := os.Getenv("PGUSER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("PGPASSWORD")
	sslmode := os.Getenv("PGSSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, name, sslmode), nil
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
	)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
