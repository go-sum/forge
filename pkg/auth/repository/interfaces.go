// Package repository defines the storage ports that the auth package requires.
package repository

import (
	"context"
	"time"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
)

// UserReader exposes read-only user lookup for middleware and auth flows.
type UserReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
}

// UserStore is the narrow user storage interface required by auth flows.
type UserStore interface {
	UserReader
	Create(ctx context.Context, email, displayName, role string, verified bool) (model.User, error)
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) (model.User, error)
	SetWebAuthnID(ctx context.Context, id uuid.UUID, webauthnID []byte) (model.User, error)
	SetWebAuthnIDIfNull(ctx context.Context, id uuid.UUID, webauthnID []byte) (model.User, error)
	GetByWebAuthnID(ctx context.Context, webauthnID []byte) (model.User, error)
}

// AdminStore defines data access operations for admin user management.
// Auth-related operations (Create, GetByEmail, UpdateEmail) are owned by
// UserStore and are intentionally absent here.
type AdminStore interface {
	UserReader
	List(ctx context.Context, limit, offset int32) ([]model.User, error)
	Update(ctx context.Context, id uuid.UUID, email, displayName, role string) (model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	HasAdmin(ctx context.Context) (bool, error)
}

// PasskeyCredentialStore defines persistence operations for WebAuthn credentials.
type PasskeyCredentialStore interface {
	CreateCredential(ctx context.Context, cred model.PasskeyCredential) (model.PasskeyCredential, error)
	GetByCredentialID(ctx context.Context, credentialID []byte) (model.PasskeyCredential, error)
	GetByIDForUser(ctx context.Context, userID, id uuid.UUID) (model.PasskeyCredential, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.PasskeyCredential, error)
	TouchPasskeyCredential(ctx context.Context, id uuid.UUID, signCount int64, cloneWarning bool, lastUsed time.Time) error
	RenameCredential(ctx context.Context, id, userID uuid.UUID, name string) (model.PasskeyCredential, error)
	DeleteCredential(ctx context.Context, id, userID uuid.UUID) error
}
