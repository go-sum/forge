// Package repository defines the storage ports that the auth package requires.
package repository

import (
	"context"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
)

// UserReader is the narrow read-only user storage interface required by auth.
type UserReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
}

// UserWriter extends UserReader with creation for use in registration transactions.
type UserWriter interface {
	UserReader
	Create(ctx context.Context, email, displayName, role string) (model.User, error)
}

// PasswordStore is the narrow password storage interface required by auth.
type PasswordStore interface {
	Create(ctx context.Context, userID uuid.UUID, hash string) (model.Password, error)
	GetCurrentByEmail(ctx context.Context, email string) (model.Password, error)
	GetCurrentByUserID(ctx context.Context, userID uuid.UUID) (model.Password, error)
}

// TxRepos holds transaction-scoped repositories for use within a single transaction.
type TxRepos struct {
	User     UserWriter
	Password PasswordStore
}
