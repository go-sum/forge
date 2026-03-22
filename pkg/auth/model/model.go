// Package model defines auth domain types: user identity, credentials, and auth errors.
package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Role constants for user access levels.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User is the auth module's view of an application user.
// It contains only the fields required for authentication and session management.
type User struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Password is the domain representation of a stored password record.
type Password struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Hash      string
	CreatedAt time.Time
}

// LoginInput carries validated credentials for authentication.
type LoginInput struct {
	Email    string `form:"email"    validate:"required,email"`
	Password string `form:"password" validate:"required"`
}

// CreateUserInput carries validated data for registering a new user.
// Role defaults to "user" when empty (applied by the service layer).
type CreateUserInput struct {
	Email       string `form:"email"         validate:"required,email,max=255"`
	DisplayName string `form:"display_name"  validate:"required,min=1,max=255"`
	Password    string `form:"password"      validate:"required,min=8"`
	Role        string `form:"role"          validate:"omitempty,oneof=user admin"`
}

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
