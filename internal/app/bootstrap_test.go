package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-sum/forge/internal/health"
	"github.com/jackc/pgx/v5"
)

type fakeRow struct {
	values []string
	err    error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for i, v := range r.values {
		s, ok := dest[i].(*string)
		if !ok {
			return errors.New("destination must be *string")
		}
		*s = v
	}
	return nil
}

type fakeQuerier struct {
	row pgx.Row
}

func (q fakeQuerier) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return q.row
}

func TestVerifyRequiredRelations(t *testing.T) {
	t.Run("accepts applied schema", func(t *testing.T) {
		err := health.VerifyRequiredRelations(context.Background(), fakeQuerier{
			row: fakeRow{values: []string{"users"}},
		})
		if err != nil {
			t.Fatalf("VerifyRequiredRelations() error = %v", err)
		}
	})

	t.Run("reports missing relations with remediation", func(t *testing.T) {
		err := health.VerifyRequiredRelations(context.Background(), fakeQuerier{
			row: fakeRow{values: []string{""}},
		})
		if err == nil {
			t.Fatal("VerifyRequiredRelations() error = nil, want missing relation error")
		}
		if !strings.Contains(err.Error(), "users") {
			t.Fatalf("error %q did not mention missing users table", err)
		}
		if !strings.Contains(err.Error(), "make db-apply") {
			t.Fatalf("error %q did not mention remediation", err)
		}
	})
}
