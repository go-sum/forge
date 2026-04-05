package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/go-sum/forge/internal/model"
	"github.com/jackc/pgx/v5"
)

func TestVerifyRequiredRelations(t *testing.T) {
	t.Run("passes when users relation exists", func(t *testing.T) {
		err := VerifyRequiredRelations(context.Background(), stubDBTX{
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row {
				return stubRow{values: []any{"users"}}
			},
		})
		if err != nil {
			t.Fatalf("VerifyRequiredRelations() error = %v", err)
		}
	})

	t.Run("reports missing relations with remediation", func(t *testing.T) {
		err := VerifyRequiredRelations(context.Background(), stubDBTX{
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row {
				return stubRow{values: []any{""}}
			},
		})
		if err == nil {
			t.Fatal("VerifyRequiredRelations() error = nil")
		}
		if !errors.Is(err, model.ErrRequiredRelationsMissing) {
			t.Fatalf("errors.Is(err, model.ErrRequiredRelationsMissing) = false; err = %v", err)
		}
		want := "required relations missing: missing required relations users; run `make db-migrate` to apply pending migrations"
		if err.Error() != want {
			t.Fatalf("VerifyRequiredRelations() error = %q, want %q", err.Error(), want)
		}
	})
}
