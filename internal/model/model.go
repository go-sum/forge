// Package model defines the domain types used across the service, repository,
// and handler layers. These types are intentionally decoupled from the db
// package's sqlc-generated structs.
package model

import (
	"time"

	authmodel "github.com/go-sum/auth/model"
	"github.com/google/uuid"
)

const (
	RoleUser  = authmodel.RoleUser
	RoleAdmin = authmodel.RoleAdmin
)

// User is the application's domain type for an authenticated user.
// It mirrors authmodel.User field-for-field so the mapping is zero-cost,
// but ownership lives here — internal/ code never imports authmodel.User directly.
type User struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	Role        string
	Verified    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserFromAuth converts an auth-package user to the application domain User.
// Use this at adapter boundaries where pkg/auth returns authmodel.User.
func UserFromAuth(u authmodel.User) User {
	return User{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		Verified:    u.Verified,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// UpdateUserInput carries validated data for updating an existing user.
// Empty strings are treated as "no change" by the COALESCE logic in the SQL query.
type UpdateUserInput struct {
	Email       string `form:"email"         validate:"omitempty,email,max=255"`
	DisplayName string `form:"display_name"  validate:"omitempty,max=255"`
	Role        string `form:"role"          validate:"omitempty,oneof=user admin"`
}
