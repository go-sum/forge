// Package migrate wraps the goose library to run PostgreSQL schema migrations.
//
// Each function opens a temporary database/sql connection (via the pgx stdlib
// driver), performs the requested operation, and closes the connection. This
// keeps migration concerns separate from the pgxpool used by the application.
package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // register pgx as database/sql driver
	"github.com/pressly/goose/v3"
)

// MigrationStatus describes the state of a single migration.
type MigrationStatus struct {
	Version   int64
	Applied   bool
	AppliedAt time.Time
	Source    string
}

// Up opens a temporary database/sql connection, runs all pending
// goose migrations from migrationsDir, then closes the connection.
func Up(ctx context.Context, dsn string, migrationsDir string) error {
	db, err := openDB(dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrate up: set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

// Status returns the state of each known migration.
func Status(ctx context.Context, dsn string, migrationsDir string) ([]MigrationStatus, error) {
	db, err := openDB(dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("migrate status: set dialect: %w", err)
	}

	migrations, err := goose.CollectMigrations(migrationsDir, 0, goose.MaxVersion)
	if err != nil {
		return nil, fmt.Errorf("migrate status: collect: %w", err)
	}

	current, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate status: db version: %w", err)
	}

	var statuses []MigrationStatus
	for _, m := range migrations {
		s := MigrationStatus{
			Version: m.Version,
			Source:  m.Source,
			Applied: m.Version <= current,
		}
		statuses = append(statuses, s)
	}

	return statuses, nil
}

// Version returns the current migration version (0 if none applied).
func Version(ctx context.Context, dsn string) (int64, error) {
	db, err := openDB(dsn)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return 0, fmt.Errorf("migrate version: set dialect: %w", err)
	}

	v, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("migrate version: %w", err)
	}

	return v, nil
}

// Down rolls back the most recently applied migration.
func Down(ctx context.Context, dsn string, migrationsDir string) error {
	db, err := openDB(dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrate down: set dialect: %w", err)
	}

	if err := goose.DownContext(ctx, db, migrationsDir); err != nil {
		return fmt.Errorf("migrate down: %w", err)
	}

	return nil
}

// Create generates a new empty SQL migration file in migrationsDir with the
// given name. It returns the path of the created file.
func Create(migrationsDir string, name string) (string, error) {
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		return "", fmt.Errorf("migrate create: mkdir: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return "", fmt.Errorf("migrate create: set dialect: %w", err)
	}

	if err := goose.Create(nil, migrationsDir, name, "sql"); err != nil {
		return "", fmt.Errorf("migrate create: %w", err)
	}

	// Find the created file (goose names it with a timestamp prefix)
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return "", fmt.Errorf("migrate create: read dir: %w", err)
	}

	// Return the last .sql file — goose creates sequential timestamps
	var lastSQL string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			lastSQL = filepath.Join(migrationsDir, e.Name())
		}
	}

	return lastSQL, nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("migrate: open db: %w", err)
	}

	db.SetMaxOpenConns(1)

	return db, nil
}
