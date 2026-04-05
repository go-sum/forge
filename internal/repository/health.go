package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-sum/forge/internal/model"

	"github.com/jackc/pgx/v5"
)

// DBQuerier is satisfied by pgx pools and test fakes.
type DBQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// VerifyRequiredRelations checks that the app's required relations exist.
func VerifyRequiredRelations(ctx context.Context, db DBQuerier) error {
	var usersTable string
	err := db.QueryRow(ctx, `
SELECT
    COALESCE(to_regclass('public.users')::text, '')
`).Scan(&usersTable)
	if err != nil {
		return fmt.Errorf("verify schema: %w", err)
	}

	missing := make([]string, 0, 1)
	if usersTable == "" {
		missing = append(missing, "users")
	}
	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf(
		"%w: missing required relations %s; run `make db-migrate` to apply pending migrations",
		model.ErrRequiredRelationsMissing,
		strings.Join(missing, ", "),
	)
}
