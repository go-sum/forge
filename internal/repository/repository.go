// Package repository provides the data-access layer. Each type wraps
// sqlc-generated queries and translates between db structs and domain models.
// Service layer code uses the interfaces defined here; implementations are
// package-private.
package repository

import (
	"context"

	"starter/internal/model"
	db "starter/db/schema"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository defines data access operations for users.
type UserRepository interface {
	Create(ctx context.Context, email, displayName, role string) (model.User, error)
	CreateWithTx(ctx context.Context, tx pgx.Tx, email, displayName, role string) (model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
	List(ctx context.Context, limit, offset int32) ([]model.User, error)
	Update(ctx context.Context, id uuid.UUID, email, displayName, role string) (model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// PasswordRepository defines data access operations for password records.
type PasswordRepository interface {
	Create(ctx context.Context, userID uuid.UUID, hash string) (model.Password, error)
	CreateWithTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, hash string) (model.Password, error)
	GetCurrentByUserID(ctx context.Context, userID uuid.UUID) (model.Password, error)
	GetCurrentByEmail(ctx context.Context, email string) (model.Password, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.Password, error)
}

// Repositories is the composition root for all data access.
// Use WithTx to obtain a transaction-scoped copy.
type Repositories struct {
	User     UserRepository
	Password PasswordRepository
	pool     *pgxpool.Pool
}

// NewRepositories constructs Repositories backed by the given pool.
func NewRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		User:     newUserRepository(pool),
		Password: newPasswordRepository(pool),
		pool:     pool,
	}
}

// WithTx returns a new Repositories where both repos share the given transaction.
// Used by AuthService.Register to atomically create a user and their password.
func (r *Repositories) WithTx(tx pgx.Tx) *Repositories {
	q := db.New(tx)
	return &Repositories{
		User:     &userRepository{q: q},
		Password: &passwordRepository{q: q},
	}
}
