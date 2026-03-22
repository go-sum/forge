// Package model defines the domain types used across the service, repository,
// and handler layers. These types are intentionally decoupled from the db
// package's sqlc-generated structs.
package model

import (
	"time"

	"github.com/google/uuid"
)

// Role constants for user access levels. Struct tag validate strings (e.g.
// `validate:"oneof=user admin"`) cannot reference Go constants — those literals
// are kept in sync with these values.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User is the domain representation of an application user.
type User struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Password is the domain representation of a stored password record.
// Passwords are append-only; the current password is the most recent row.
type Password struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Hash      string
	CreatedAt time.Time
}

// UpdateUserInput carries validated data for updating an existing user.
// Empty strings are treated as "no change" by the COALESCE logic in the SQL query.
type UpdateUserInput struct {
	Email       string `form:"email"         validate:"omitempty,email,max=255"`
	DisplayName string `form:"display_name"  validate:"omitempty,max=255"`
	Role        string `form:"role"          validate:"omitempty,oneof=user admin"`
}
