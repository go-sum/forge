// Package database provides PostgreSQL connection pool management via pgxpool.
//
// MaxConns defaults to 10 when not specified in the DSN. PostgreSQL's default
// max_connections is 100; with multiple app instances the dynamic pgxpool
// default (max(4, GOMAXPROCS*4)) can exhaust that silently. Override via DSN
// query parameter: ?pool_max_conns=20.
package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a new connection pool and verifies connectivity.
// maxConns sets pool_max_conns; when 0 or negative the DSN value is used,
// falling back to 10 if the DSN also omits it.
func Connect(ctx context.Context, dsn string, maxConns int32) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing database config: %w", err)
	}

	// Apply a safe MaxConns cap. pgxpool's dynamic default is
	// max(4, GOMAXPROCS*4), which on a 4-core host gives 16 connections.
	// With multiple app instances this exhausts PostgreSQL's default
	// max_connections=100 silently. The DSN pool_max_conns=N param takes
	// precedence because pgxpool.ParseConfig already wrote it into poolCfg.
	if maxConns > 0 {
		poolCfg.MaxConns = maxConns
	} else if poolCfg.MaxConns == 0 {
		poolCfg.MaxConns = 10
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return pool, nil
}

// IsUniqueViolation reports whether err is a PostgreSQL unique constraint
// violation (SQLSTATE 23505). Use this to map database errors to domain errors
// without duplicating the postgres error code string at each call site.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
