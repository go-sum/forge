package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	pgplan "github.com/pgplex/pgschema/cmd/plan"
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
			"composed desired-state file, generates a pgschema migration plan, then\n" +
			"writes a goose migration file with -- +goose Up/Down sections.",
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

	// 2. Compose all schema files into a single temp file for pgschema.
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

	// 3. Build plan config from standard PostgreSQL environment variables.
	config, err := buildPlanConfig(tmpFile.Name())
	if err != nil {
		return err
	}

	// 4. Create desired-state provider and generate migration plan.
	provider, err := pgplan.CreateDesiredStateProvider(config)
	if err != nil {
		return fmt.Errorf("compose: create provider: %w", err)
	}
	defer provider.Stop()

	migrationPlan, err := pgplan.GeneratePlan(config, provider)
	if err != nil {
		return fmt.Errorf("compose: generate plan: %w", err)
	}

	// 5. Print human-readable summary to terminal (always coloured).
	fmt.Println("\n── Schema Diff ──────────────────────────────────────────────")
	fmt.Print(migrationPlan.HumanColored(true))

	if !migrationPlan.HasAnyChanges() {
		fmt.Println("\nNo schema changes detected.")
		return nil
	}

	if diffOnly {
		return nil
	}

	// 6. Extract Up SQL from the plan's JSON representation.
	upSQL, err := extractUpSQL(migrationPlan)
	if err != nil {
		return fmt.Errorf("compose: extract SQL: %w", err)
	}

	// 7. Generate Down SQL by reversing the Up statements.
	downSQL := buildDownSQL(upSQL)

	// 8. Create and write the goose migration file.
	path, err := migrate.Create(dir, name)
	if err != nil {
		return fmt.Errorf("compose: create migration: %w", err)
	}

	content := "-- +goose Up\n" + upSQL + "\n-- +goose Down\n" + downSQL
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("compose: write migration: %w", err)
	}

	fmt.Printf("\nCreated migration: %s\n", path)
	fmt.Println("Review and edit the migration file before running db-migrate.")
	return nil
}

// buildPlanConfig constructs a PlanConfig from standard PostgreSQL environment variables.
// The plan database follows the naming convention established in db/init-dev/01-databases.sh:
// it lives on the same host/port/user/password as the target DB, with a "_plan" suffix.
func buildPlanConfig(schemaFile string) (*pgplan.PlanConfig, error) {
	port := 5432
	if p := os.Getenv("PGPORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("compose: invalid PGPORT %q: %w", p, err)
		}
		port = n
	}

	host := os.Getenv("PGHOST")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	db := os.Getenv("PGDATABASE")

	return &pgplan.PlanConfig{
		Host:            host,
		Port:            port,
		DB:              db,
		User:            user,
		Password:        password,
		Schema:          "public",
		File:            schemaFile,
		ApplicationName: "forge-db-compose",
		SSLMode:         "prefer",
		// Plan DB: same connection details, database name suffixed with "_plan".
		// Matches the database created by db/init-dev/01-databases.sh.
		PlanDBHost:     host,
		PlanDBPort:     port,
		PlanDBDatabase: db + "_plan",
		PlanDBUser:     user,
		PlanDBPassword: password,
		PlanDBSSLMode:  "prefer",
	}, nil
}

// planJSON is a minimal struct for unmarshalling the SQL statements from plan.Plan.ToJSON().
// Field names match the json tags on internal/plan.Plan, ExecutionGroup, and Step.
type planJSON struct {
	Groups []struct {
		Steps []struct {
			SQL string `json:"sql"`
		} `json:"steps"`
	} `json:"groups"`
}

// planResult is a local interface over *internal/plan.Plan so we can call its methods
// without importing the internal package directly.
type planResult interface {
	ToJSON() (string, error)
	HumanColored(enableColor bool) string
	HasAnyChanges() bool
}

func extractUpSQL(p planResult) (string, error) {
	raw, err := p.ToJSON()
	if err != nil {
		return "", fmt.Errorf("marshal plan: %w", err)
	}

	var out planJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", fmt.Errorf("parse plan JSON: %w", err)
	}

	var sb strings.Builder
	for _, g := range out.Groups {
		for _, step := range g.Steps {
			stmt := strings.TrimSpace(step.SQL)
			if stmt == "" {
				continue
			}
			// goose_db_version is goose's internal tracking table; it is created
			// and managed by goose itself and must never appear in a migration file.
			if strings.Contains(stmt, "goose_db_version") {
				continue
			}
			sb.WriteString(stmt)
			sb.WriteString("\n\n")
		}
	}
	return strings.TrimRight(sb.String(), "\n") + "\n", nil
}

// Patterns for reversing DDL statements into their Down counterparts.
var (
	reCreateTable   = regexp.MustCompile(`(?i)^CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\S+)`)
	reCreateIndex   = regexp.MustCompile(`(?i)^CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?(\S+)`)
	reCreateTrigger = regexp.MustCompile(`(?is)^CREATE\s+(?:OR\s+REPLACE\s+)?TRIGGER\s+(\S+).*?\bON\s+(\S+)`)
	reAlterAddCol   = regexp.MustCompile(`(?i)^ALTER\s+TABLE\s+(\S+)\s+ADD\s+COLUMN\s+(\S+)`)
)

// buildDownSQL generates Down SQL by reversing each Up statement.
// Statements are processed in reverse order so the last thing created is the first dropped.
// Unrecognised or already-DROP statements are silently skipped.
func buildDownSQL(upSQL string) string {
	stmts := splitStatements(upSQL)
	drops := make([]string, 0, len(stmts))

	for i := len(stmts) - 1; i >= 0; i-- {
		stmt := strings.TrimSpace(stmts[i])
		if stmt == "" {
			continue
		}
		firstLine := strings.SplitN(stmt, "\n", 2)[0]

		switch {
		case reCreateTable.MatchString(firstLine):
			m := reCreateTable.FindStringSubmatch(firstLine)
			drops = append(drops, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", m[1]))
		case reCreateIndex.MatchString(firstLine):
			m := reCreateIndex.FindStringSubmatch(firstLine)
			drops = append(drops, fmt.Sprintf("DROP INDEX IF EXISTS %s;", m[1]))
		case reCreateTrigger.MatchString(stmt):
			m := reCreateTrigger.FindStringSubmatch(stmt)
			drops = append(drops, fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s;", m[1], m[2]))
		case reAlterAddCol.MatchString(firstLine):
			m := reAlterAddCol.FindStringSubmatch(firstLine)
			drops = append(drops, fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s;", m[1], m[2]))
		}
	}

	if len(drops) == 0 {
		return ""
	}
	return strings.Join(drops, "\n") + "\n"
}

// splitStatements splits a SQL string on the statement-terminating semicolons
// produced by pgschema's SQL output format.
func splitStatements(sql string) []string {
	raw := strings.Split(sql, ";\n")
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s+";")
		}
	}
	return out
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
