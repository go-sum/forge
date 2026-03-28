// Package model defines the domain types used across the service, repository,
// and handler layers. These types are intentionally decoupled from the db
// package's sqlc-generated structs.
package model

import (
	authmodel "github.com/go-sum/auth/model"
)

const (
	RoleUser  = authmodel.RoleUser
	RoleAdmin = authmodel.RoleAdmin
)

type User = authmodel.User

// UpdateUserInput carries validated data for updating an existing user.
// Empty strings are treated as "no change" by the COALESCE logic in the SQL query.
type UpdateUserInput struct {
	Email       string `form:"email"         validate:"omitempty,email,max=255"`
	DisplayName string `form:"display_name"  validate:"omitempty,max=255"`
	Role        string `form:"role"          validate:"omitempty,oneof=user admin"`
}
