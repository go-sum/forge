package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-sum/server/database/migrate"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

type schemasConfig struct {
	Schemas []string `yaml:"schemas"`
}

func newComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compose <name>",
		Short: "Compose schemas and generate a migration with diff",
		Long: "Reads db/sql/schemas.yaml, concatenates all schema files into a\n" +
			"composed desired-state file, creates a goose migration, then runs\n" +
			"pgschema plan to show the schema diff.",
		Args: cobra.ExactArgs(1),
		RunE: runCompose,
	}
	cmd.Flags().String("dir", "db/migrations", "migrations directory")
	cmd.Flags().String("config", "db/sql/schemas.yaml", "schema registry file")
	cmd.Flags().Bool("diff-only", false, "show diff without creating a migration file")
	return cmd
}

func runCompose(cmd *cobra.Command, args []string) error {
	name := args[0]
	dir, _ := cmd.Flags().GetString("dir")
	configPath, _ := cmd.Flags().GetString("config")
	diffOnly, _ := cmd.Flags().GetBool("diff-only")

	// 1. Parse schemas.yaml
	schemas, err := loadSchemas(configPath)
	if err != nil {
		return err
	}
	fmt.Printf("Composing %d schema source(s):\n", len(schemas))
	for _, s := range schemas {
		fmt.Printf("  - %s\n", s)
	}

	// 2. Compose into a temp file
	composed, err := composeSchemas(schemas)
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp("", "composed-schema-*.sql")
	if err != nil {
		return fmt.Errorf("compose: create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(composed); err != nil {
		tmpFile.Close()
		return fmt.Errorf("compose: write temp file: %w", err)
	}
	tmpFile.Close()

	// 3. Create migration file (unless diff-only)
	if !diffOnly {
		path, err := migrate.Create(dir, name)
		if err != nil {
			return fmt.Errorf("compose: create migration: %w", err)
		}
		fmt.Printf("\nCreated migration: %s\n", path)
	}

	// 4. Run pgschema plan
	fmt.Println("\n── Schema Diff ──────────────────────────────────────────────")
	ctx := context.Background()
	pgCmd := exec.CommandContext(ctx, "pgschema", "plan",
		"--file", tmpFile.Name(),
		"--output-human", "stdout",
	)
	pgCmd.Stdout = os.Stdout
	pgCmd.Stderr = os.Stderr

	if err := pgCmd.Run(); err != nil {
		return fmt.Errorf("compose: pgschema plan: %w", err)
	}

	if !diffOnly {
		fmt.Println("\nReview the diff above, then edit the migration file in", dir)
	}
	return nil
}

func loadSchemas(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("compose: read config %s: %w", path, err)
	}

	var cfg schemasConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("compose: parse config %s: %w", path, err)
	}

	if len(cfg.Schemas) == 0 {
		return nil, fmt.Errorf("compose: no schemas listed in %s", path)
	}

	return cfg.Schemas, nil
}

func composeSchemas(paths []string) (string, error) {
	var buf strings.Builder
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return "", fmt.Errorf("compose: read schema %s: %w", p, err)
		}
		fmt.Fprintf(&buf, "-- ── Source: %s ──\n", p)
		buf.Write(data)
		buf.WriteByte('\n')
	}
	return buf.String(), nil
}
