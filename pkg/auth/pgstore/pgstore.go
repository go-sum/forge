// Package pgstore implements the auth repository.UserStore interface using
// PostgreSQL with pgx/v5. It owns the users table schema and all auth-related
// queries. Call Install() once at startup to create the table idempotently.
package pgstore

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/go-sum/auth/model"
	authdb "github.com/go-sum/auth/pgstore/db"
	authrepo "github.com/go-sum/auth/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Compile-time interface check.
var _ authrepo.UserStore = (*PgStore)(nil)

//go:embed sql/schema.sql
var createTableSQL string

// Config holds the PostgreSQL store configuration.
type Config struct {
	Pool *pgxpool.Pool
}

// PgStore implements authrepo.UserStore backed by PostgreSQL.
type PgStore struct {
	pool    *pgxpool.Pool
	queries *authdb.Queries
}

// New creates a PgStore. The pool is externally managed and not closed by PgStore.
func New(cfg Config) *PgStore {
	return &PgStore{
		pool:    cfg.Pool,
		queries: authdb.New(cfg.Pool),
	}
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
	u, err := s.queries.CreateUser(ctx, authdb.CreateUserParams{
		Email:       email,
		DisplayName: displayName,
		Role:        role,
		Verified:    verified,
	})
	if err != nil {
		return model.User{}, mapUserErr(err)
	}
	return toUserModel(u), nil
}

// GetByID retrieves a user by UUID. Returns model.ErrUserNotFound when missing.
func (s *PgStore) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	u, err := s.queries.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return toUserModel(u), nil
}

// GetByEmail retrieves a user by email address (case-insensitive via CITEXT).
// Returns model.ErrUserNotFound when missing.
func (s *PgStore) GetByEmail(ctx context.Context, email string) (model.User, error) {
	u, err := s.queries.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return toUserModel(u), nil
}

// UpdateEmail changes the user's email address and returns the updated record.
// Returns model.ErrUserNotFound when no row matches id.
// Returns model.ErrEmailTaken on unique constraint violation.
func (s *PgStore) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (model.User, error) {
	u, err := s.queries.UpdateUserEmail(ctx, authdb.UpdateUserEmailParams{
		ID:    id,
		Email: email,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, mapUserErr(err)
	}
	return toUserModel(u), nil
}

// toUserModel converts a sqlc-generated db.User to the domain model.
func toUserModel(u authdb.User) model.User {
	return model.User{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		Verified:    u.Verified,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// mapUserErr translates PostgreSQL unique constraint violations to domain errors.
func mapUserErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return model.ErrEmailTaken
	}
	return err
}
