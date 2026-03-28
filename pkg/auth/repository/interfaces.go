// Package repository defines the storage ports that the auth package requires.
package repository

import (
	"context"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
)

// UserReader exposes read-only user lookup for middleware and auth flows.
type UserReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
}

// UserStore is the narrow user storage interface required by auth.
type UserStore interface {
	UserReader
	Create(ctx context.Context, email, displayName, role string, verified bool) (model.User, error)
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) (model.User, error)
}
