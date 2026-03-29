package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/go-sum/forge/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	repoTestUserID    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repoTestCreatedAt = time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	repoTestUpdatedAt = time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
)

type stubDBTX struct {
	execFn     func(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	queryFn    func(context.Context, string, ...interface{}) (pgx.Rows, error)
	queryRowFn func(context.Context, string, ...interface{}) pgx.Row
}

func (s stubDBTX) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if s.execFn != nil {
		return s.execFn(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, nil
}

func (s stubDBTX) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if s.queryFn != nil {
		return s.queryFn(ctx, sql, args...)
	}
	return &stubRows{}, nil
}

func (s stubDBTX) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if s.queryRowFn != nil {
		return s.queryRowFn(ctx, sql, args...)
	}
	return stubRow{}
}

type stubRow struct {
	values []any
	err    error
}

func (r stubRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for i, v := range r.values {
		if err := assignScanValue(dest[i], v); err != nil {
			return err
		}
	}
	return nil
}

type stubRows struct {
	rows [][]any
	idx  int
	err  error
}

func (r stubRows) Close() {}

func (r stubRows) Err() error { return r.err }

func (r stubRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

func (r stubRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

func (r *stubRows) Next() bool {
	if r.idx >= len(r.rows) {
		return false
	}
	r.idx++
	return true
}

func (r stubRows) Scan(dest ...any) error {
	row := r.rows[r.idx-1]
	for i, v := range row {
		if err := assignScanValue(dest[i], v); err != nil {
			return err
		}
	}
	return nil
}

func (r stubRows) Values() ([]any, error) { return nil, nil }

func (r stubRows) RawValues() [][]byte { return nil }

func (r stubRows) Conn() *pgx.Conn { return nil }

func assignScanValue(dest, value any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("scan dest must be a non-nil pointer")
	}
	valueV := reflect.ValueOf(value)
	if !valueV.Type().AssignableTo(rv.Elem().Type()) {
		return errors.New("scan value type mismatch")
	}
	rv.Elem().Set(valueV)
	return nil
}

func TestUserRepositoryMethods(t *testing.T) {
	ctx := context.Background()
	expected := model.User{
		ID:          repoTestUserID,
		Email:       "ada@example.com",
		DisplayName: "Ada Lovelace",
		Role:        "admin",
		Verified:    true,
		CreatedAt:   repoTestCreatedAt,
		UpdatedAt:   repoTestUpdatedAt,
	}

	t.Run("create and get map db users", func(t *testing.T) {
		repo := newUserRepository(stubDBTX{
			queryRowFn: func(_ context.Context, _ string, args ...interface{}) pgx.Row {
				switch len(args) {
				case 4:
					if args[0] != expected.Email || args[1] != expected.DisplayName || args[2] != expected.Role || args[3] != true {
						t.Fatalf("args = %#v", args)
					}
				case 1:
					if args[0] != repoTestUserID && args[0] != expected.Email {
						t.Fatalf("args = %#v", args)
					}
				}
				return stubRow{values: []any{
					expected.ID, expected.Email, expected.DisplayName, expected.Role, expected.Verified, expected.CreatedAt, expected.UpdatedAt,
				}}
			},
		})

		created, err := repo.Create(ctx, expected.Email, expected.DisplayName, expected.Role, true)
		if err != nil || created != expected {
			t.Fatalf("Create() user=%#v err=%v", created, err)
		}

		byID, err := repo.GetByID(ctx, repoTestUserID)
		if err != nil || byID != expected {
			t.Fatalf("GetByID() user=%#v err=%v", byID, err)
		}

		byEmail, err := repo.GetByEmail(ctx, expected.Email)
		if err != nil || byEmail != expected {
			t.Fatalf("GetByEmail() user=%#v err=%v", byEmail, err)
		}
	})

	t.Run("get methods map no rows", func(t *testing.T) {
		repo := newUserRepository(stubDBTX{
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row { return stubRow{err: pgx.ErrNoRows} },
		})
		if _, err := repo.GetByID(ctx, repoTestUserID); !errors.Is(err, model.ErrUserNotFound) {
			t.Fatalf("GetByID() err = %v", err)
		}
		if _, err := repo.GetByEmail(ctx, expected.Email); !errors.Is(err, model.ErrUserNotFound) {
			t.Fatalf("GetByEmail() err = %v", err)
		}
	})

	t.Run("list update delete count and error mapping", func(t *testing.T) {
		repo := newUserRepository(stubDBTX{
			queryFn: func(_ context.Context, _ string, args ...interface{}) (pgx.Rows, error) {
				if args[0] != int32(10) || args[1] != int32(20) {
					t.Fatalf("args = %#v", args)
				}
				rows := &stubRows{rows: [][]any{{
					expected.ID, expected.Email, expected.DisplayName, expected.Role, expected.Verified, expected.CreatedAt, expected.UpdatedAt,
				}}}
				return rows, nil
			},
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row {
				return stubRow{values: []any{
					expected.ID, expected.Email, expected.DisplayName, expected.Role, expected.Verified, expected.CreatedAt, expected.UpdatedAt,
				}}
			},
			execFn: func(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
		})

		users, err := repo.List(ctx, 10, 20)
		if err != nil || len(users) != 1 || users[0] != expected {
			t.Fatalf("List() users=%#v err=%v", users, err)
		}

		updated, err := repo.Update(ctx, repoTestUserID, "", "New Name", "")
		if err != nil || updated != expected {
			t.Fatalf("Update() user=%#v err=%v", updated, err)
		}

		if err := repo.Delete(ctx, repoTestUserID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		countRepo := newUserRepository(stubDBTX{
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row { return stubRow{values: []any{int64(7)}} },
		})
		count, err := countRepo.Count(ctx)
		if err != nil || count != 7 {
			t.Fatalf("Count() count=%d err=%v", count, err)
		}
	})

	t.Run("create and update map unique violation and no rows", func(t *testing.T) {
		uniqueErr := &pgconn.PgError{Code: "23505"}
		createRepo := newUserRepository(stubDBTX{
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row { return stubRow{err: uniqueErr} },
		})
		if _, err := createRepo.Create(ctx, expected.Email, expected.DisplayName, expected.Role, true); !errors.Is(err, model.ErrEmailTaken) {
			t.Fatalf("Create() err = %v", err)
		}

		updateNoRows := newUserRepository(stubDBTX{
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row { return stubRow{err: pgx.ErrNoRows} },
		})
		if _, err := updateNoRows.Update(ctx, repoTestUserID, "", "", ""); !errors.Is(err, model.ErrUserNotFound) {
			t.Fatalf("Update() err = %v", err)
		}

		updateUnique := newUserRepository(stubDBTX{
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row { return stubRow{err: uniqueErr} },
		})
		if _, err := updateUnique.Update(ctx, repoTestUserID, expected.Email, "", ""); !errors.Is(err, model.ErrEmailTaken) {
			t.Fatalf("Update() err = %v", err)
		}
	})
}

func TestRepositoriesConstructors(t *testing.T) {
	repos := NewRepositories(nil)
	if repos.User == nil {
		t.Fatalf("NewRepositories() = %#v", repos)
	}
}
