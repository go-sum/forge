package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"starter/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	repoTestUserID     = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repoTestPasswordID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repoTestCreatedAt  = time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	repoTestUpdatedAt  = time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
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

type stubTx struct {
	stubDBTX
}

func (tx stubTx) Begin(context.Context) (pgx.Tx, error) { return tx, nil }

func (stubTx) Commit(context.Context) error { return nil }

func (stubTx) Rollback(context.Context) error { return nil }

func (stubTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("unexpected CopyFrom call")
}

func (stubTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }

func (stubTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

func (stubTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("unexpected Prepare call")
}

func (tx stubTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return tx.stubDBTX.Exec(ctx, sql, args...)
}

func (tx stubTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return tx.stubDBTX.Query(ctx, sql, args...)
}

func (tx stubTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return tx.stubDBTX.QueryRow(ctx, sql, args...)
}

func (stubTx) Conn() *pgx.Conn { return nil }

func assignScanValue(dest, value any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
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
		CreatedAt:   repoTestCreatedAt,
		UpdatedAt:   repoTestUpdatedAt,
	}

	t.Run("create and get map db users", func(t *testing.T) {
		repo := newUserRepository(stubDBTX{
			queryRowFn: func(_ context.Context, _ string, args ...interface{}) pgx.Row {
				switch len(args) {
				case 3:
					if args[0] != expected.Email || args[1] != expected.DisplayName || args[2] != expected.Role {
						t.Fatalf("args = %#v", args)
					}
				case 1:
					if args[0] != repoTestUserID && args[0] != expected.Email {
						t.Fatalf("args = %#v", args)
					}
				}
				return stubRow{values: []any{
					expected.ID, expected.Email, expected.DisplayName, expected.Role, expected.CreatedAt, expected.UpdatedAt,
				}}
			},
		})

		created, err := repo.Create(ctx, expected.Email, expected.DisplayName, expected.Role)
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
					expected.ID, expected.Email, expected.DisplayName, expected.Role, expected.CreatedAt, expected.UpdatedAt,
				}}}
				return rows, nil
			},
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row {
				return stubRow{values: []any{
					expected.ID, expected.Email, expected.DisplayName, expected.Role, expected.CreatedAt, expected.UpdatedAt,
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
		if _, err := createRepo.Create(ctx, expected.Email, expected.DisplayName, expected.Role); !errors.Is(err, model.ErrEmailTaken) {
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

func TestPasswordRepositoryMethods(t *testing.T) {
	ctx := context.Background()
	expected := model.Password{
		ID:        repoTestPasswordID,
		UserID:    repoTestUserID,
		Hash:      "hashed",
		CreatedAt: repoTestCreatedAt,
	}

	t.Run("create and list map db passwords", func(t *testing.T) {
		repo := newPasswordRepository(stubDBTX{
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row {
				return stubRow{values: []any{expected.ID, expected.UserID, expected.Hash, expected.CreatedAt}}
			},
			queryFn: func(context.Context, string, ...interface{}) (pgx.Rows, error) {
				return &stubRows{rows: [][]any{{expected.ID, expected.UserID, expected.Hash, expected.CreatedAt}}}, nil
			},
		})

		created, err := repo.Create(ctx, expected.UserID, expected.Hash)
		if err != nil || created != expected {
			t.Fatalf("Create() password=%#v err=%v", created, err)
		}

		passwords, err := repo.ListByUserID(ctx, expected.UserID)
		if err != nil || len(passwords) != 1 || passwords[0] != expected {
			t.Fatalf("ListByUserID() passwords=%#v err=%v", passwords, err)
		}
	})

	t.Run("current password methods map no rows", func(t *testing.T) {
		repo := newPasswordRepository(stubDBTX{
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row { return stubRow{err: pgx.ErrNoRows} },
		})
		if _, err := repo.GetCurrentByUserID(ctx, expected.UserID); !errors.Is(err, model.ErrUserNotFound) {
			t.Fatalf("GetCurrentByUserID() err = %v", err)
		}
		if _, err := repo.GetCurrentByEmail(ctx, "ada@example.com"); !errors.Is(err, model.ErrInvalidCredentials) {
			t.Fatalf("GetCurrentByEmail() err = %v", err)
		}
	})

	t.Run("current password methods map db passwords", func(t *testing.T) {
		repo := newPasswordRepository(stubDBTX{
			queryRowFn: func(context.Context, string, ...interface{}) pgx.Row {
				return stubRow{values: []any{expected.ID, expected.UserID, expected.Hash, expected.CreatedAt}}
			},
		})
		byUser, err := repo.GetCurrentByUserID(ctx, expected.UserID)
		if err != nil || byUser != expected {
			t.Fatalf("GetCurrentByUserID() password=%#v err=%v", byUser, err)
		}
		byEmail, err := repo.GetCurrentByEmail(ctx, "ada@example.com")
		if err != nil || byEmail != expected {
			t.Fatalf("GetCurrentByEmail() password=%#v err=%v", byEmail, err)
		}
	})
}

func TestRepositoriesConstructors(t *testing.T) {
	repos := NewRepositories(nil)
	if repos.User == nil || repos.Password == nil {
		t.Fatalf("NewRepositories() = %#v", repos)
	}

	tx := stubTx{stubDBTX: stubDBTX{
		queryRowFn: func(context.Context, string, ...interface{}) pgx.Row {
			return stubRow{values: []any{repoTestPasswordID, repoTestUserID, "hashed", repoTestCreatedAt}}
		},
	}}
	txRepos := (&Repositories{}).WithTx(tx)
	if txRepos.User == nil || txRepos.Password == nil {
		t.Fatalf("WithTx() = %#v", txRepos)
	}
	password, err := txRepos.Password.Create(context.Background(), repoTestUserID, "hashed")
	if err != nil || password.UserID != repoTestUserID {
		t.Fatalf("txRepos.Password.Create() password=%#v err=%v", password, err)
	}
}
