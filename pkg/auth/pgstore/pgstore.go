// Package pgstore implements the auth repository.UserStore interface using
// PostgreSQL with pgx/v5. It owns the auth module's users-table queries; the
// host application applies schema changes via migrations composed through
// db/sql/schemas.yaml.
package pgstore

import (
	"context"
	"errors"

	"github.com/go-sum/auth/model"
	authdb "github.com/go-sum/auth/pgstore/db"
	authrepo "github.com/go-sum/auth/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Compile-time interface checks.
var (
	_ authrepo.UserStore  = (*PgStore)(nil)
	_ authrepo.AdminStore = (*PgStore)(nil)
)

// Config holds the PostgreSQL store configuration.
type Config struct {
	Pool *pgxpool.Pool
}

// PgStore implements authrepo.UserStore backed by PostgreSQL.
type PgStore struct {
	queries *authdb.Queries
}

// New creates a PgStore. The pool is externally managed and not closed by PgStore.
func New(cfg Config) *PgStore {
	return &PgStore{
		queries: authdb.New(cfg.Pool),
	}
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

// List returns a paginated slice of users, ordered by creation date descending.
func (s *PgStore) List(ctx context.Context, limit, offset int32) ([]model.User, error) {
	rows, err := s.queries.ListUsers(ctx, authdb.ListUsersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	users := make([]model.User, len(rows))
	for i, u := range rows {
		users[i] = toUserModel(u)
	}
	return users, nil
}

// Update applies non-empty fields to the user. Empty strings are treated as
// "no change" by COALESCE logic in the SQL query.
// Returns model.ErrUserNotFound when no row matches id.
// Returns model.ErrEmailTaken on unique constraint violation.
func (s *PgStore) Update(ctx context.Context, id uuid.UUID, email, displayName, role string) (model.User, error) {
	u, err := s.queries.UpdateUser(ctx, authdb.UpdateUserParams{
		ID:          id,
		Email:       email,
		DisplayName: displayName,
		Role:        role,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, mapUserErr(err)
	}
	return toUserModel(u), nil
}

// Delete removes a user. If no row matches id, the operation succeeds silently —
// DELETE is idempotent per RFC 9110 §9.3.5.
func (s *PgStore) Delete(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteUser(ctx, id)
}

// Count returns the total number of users.
func (s *PgStore) Count(ctx context.Context) (int64, error) {
	return s.queries.CountUsers(ctx)
}

// HasAdmin reports whether at least one admin user exists.
func (s *PgStore) HasAdmin(ctx context.Context) (bool, error) {
	return s.queries.HasAdminUser(ctx)
}

// SetWebAuthnID sets the WebAuthn user handle for an existing user.
// Returns model.ErrUserNotFound when no row matches id.
func (s *PgStore) SetWebAuthnID(ctx context.Context, id uuid.UUID, webauthnID []byte) (model.User, error) {
	u, err := s.queries.SetUserWebAuthnID(ctx, authdb.SetUserWebAuthnIDParams{
		ID:         id,
		WebauthnID: webauthnID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return toUserModel(u), nil
}

// SetWebAuthnIDIfNull sets the WebAuthn user handle only when the user has no handle yet.
// Returns model.ErrWebAuthnIDAlreadySet when the row already has a webauthn_id set,
// so callers can detect the "already set" case and re-read the current value.
func (s *PgStore) SetWebAuthnIDIfNull(ctx context.Context, id uuid.UUID, webauthnID []byte) (model.User, error) {
	u, err := s.queries.SetUserWebAuthnIDIfNull(ctx, authdb.SetUserWebAuthnIDIfNullParams{
		ID:         id,
		WebauthnID: webauthnID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrWebAuthnIDAlreadySet
	}
	if err != nil {
		return model.User{}, err
	}
	return toUserModel(u), nil
}

// GetByWebAuthnID retrieves a user by their WebAuthn user handle.
// Returns model.ErrUserNotFound when missing.
func (s *PgStore) GetByWebAuthnID(ctx context.Context, webauthnID []byte) (model.User, error) {
	u, err := s.queries.GetUserByWebAuthnID(ctx, webauthnID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
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
		WebAuthnID:  u.WebauthnID,
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
