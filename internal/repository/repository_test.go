package repository

import (
	"context"
	"errors"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
