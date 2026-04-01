// Package repository provides the data-access layer. Each type wraps
// sqlc-generated queries and translates between db structs and domain models.
// Service layer code uses the interfaces defined here; implementations are
// package-private.
package repository

import (
	"context"

	db "github.com/go-sum/forge/db/schema"
	"github.com/go-sum/forge/internal/model"

	"github.com/google/uuid"
)

// UserRepository defines data access operations for users.
type UserRepository interface {
	Create(ctx context.Context, email, displayName, role string, verified bool) (model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
	List(ctx context.Context, limit, offset int32) ([]model.User, error)
	Update(ctx context.Context, id uuid.UUID, email, displayName, role string) (model.User, error)
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) (model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	HasAdmin(ctx context.Context) (bool, error)
}

// Repositories is the composition root for all data access.
type Repositories struct {
	User UserRepository
}

// NewRepositories constructs Repositories backed by the given pool.
func NewRepositories(pool db.DBTX) *Repositories {
	return &Repositories{
		User: newUserRepository(pool),
	}
}
