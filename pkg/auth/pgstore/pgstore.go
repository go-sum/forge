// Package pgstore implements the auth repository.UserStore interface using
// PostgreSQL with pgx/v5. It owns the users table schema and all auth-related
// queries. Call Install() once at startup to create the table idempotently.
package pgstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-sum/auth/model"
	authrepo "github.com/go-sum/auth/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Compile-time interface check.
var _ authrepo.UserStore = (*PgStore)(nil)

// Config holds the PostgreSQL store configuration.
type Config struct {
	Pool *pgxpool.Pool
}

// PgStore implements authrepo.UserStore backed by PostgreSQL.
type PgStore struct {
	pool *pgxpool.Pool
}

// New creates a PgStore. The pool is externally managed and not closed by PgStore.
func New(cfg Config) *PgStore {
	return &PgStore{pool: cfg.Pool}
}

// Install creates the users table and indexes idempotently.
// Safe to call on an existing database — uses CREATE TABLE IF NOT EXISTS.
// Requires the citext extension to be installed by the host application.
func (s *PgStore) Install(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("pgstore: install schema: %w", err)
	}
	return nil
}

// Create inserts a new user and returns the persisted record.
func (s *PgStore) Create(ctx context.Context, email, displayName, role string, verified bool) (model.User, error) {
	row := s.pool.QueryRow(ctx, createUserSQL, email, displayName, role, verified)
	u, err := scanUser(row)
	if err != nil {
		return model.User{}, mapUserErr(err)
	}
	return u, nil
}

// GetByID retrieves a user by UUID. Returns model.ErrUserNotFound when missing.
func (s *PgStore) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	row := s.pool.QueryRow(ctx, getUserByIDSQL, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return u, nil
}

// GetByEmail retrieves a user by email address (case-insensitive via CITEXT).
// Returns model.ErrUserNotFound when missing.
func (s *PgStore) GetByEmail(ctx context.Context, email string) (model.User, error) {
	row := s.pool.QueryRow(ctx, getUserByEmailSQL, email)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return u, nil
}

// UpdateEmail changes the user's email address and returns the updated record.
// Returns model.ErrUserNotFound when no row matches id.
// Returns model.ErrEmailTaken on unique constraint violation.
func (s *PgStore) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (model.User, error) {
	row := s.pool.QueryRow(ctx, updateUserEmailSQL, id, email)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, mapUserErr(err)
	}
	return u, nil
}

// scanUser scans all user columns from a pgx.Row into a model.User.
func scanUser(row pgx.Row) (model.User, error) {
	var u model.User
	var createdAt, updatedAt time.Time
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.DisplayName,
		&u.Role,
		&u.Verified,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.User{}, err
	}
	u.CreatedAt = createdAt
	u.UpdatedAt = updatedAt
	return u, nil
}

// mapUserErr translates PostgreSQL unique constraint violations to domain errors.
func mapUserErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return model.ErrEmailTaken
	}
	return err
}
