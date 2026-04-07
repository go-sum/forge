// Example: safe to delete as a unit.
package service

import (
	"context"
	"fmt"

	authmodel "github.com/go-sum/auth/model"
	authrepo "github.com/go-sum/auth/repository"
	"github.com/go-sum/forge/internal/model"

	"github.com/google/uuid"
)

// UserService provides CRUD operations for user management.
type UserService struct {
	repo authrepo.AdminStore
}

// NewUserService constructs a UserService from an admin user store.
func NewUserService(repo authrepo.AdminStore) *UserService {
	return &UserService{repo: repo}
}

// List returns a paginated slice of users. perPage is capped at 100.
func (s *UserService) List(ctx context.Context, page, perPage int) ([]authmodel.User, error) {
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage
	return s.repo.List(ctx, int32(perPage), int32(offset))
}

// GetByID retrieves a single user by ID. Returns authmodel.ErrUserNotFound when missing.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (authmodel.User, error) {
	return s.repo.GetByID(ctx, id)
}

// Update applies non-empty fields from input to the user. Empty strings are
// treated as "no change" by the underlying COALESCE SQL logic.
func (s *UserService) Update(ctx context.Context, id uuid.UUID, input authmodel.UpdateUserInput) (authmodel.User, error) {
	return s.repo.Update(ctx, id, input.Email, input.DisplayName, input.Role)
}

// Delete removes a user by ID.
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// Count returns the total number of users.
func (s *UserService) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

// HasAdmin reports whether at least one admin user exists.
func (s *UserService) HasAdmin(ctx context.Context) (bool, error) {
	return s.repo.HasAdmin(ctx)
}

// ElevateToAdmin promotes the given user to admin, but only when no admin
// exists yet. Returns model.ErrAdminExists if an admin is already present.
func (s *UserService) ElevateToAdmin(ctx context.Context, userID uuid.UUID) (authmodel.User, error) {
	hasAdmin, err := s.repo.HasAdmin(ctx)
	if err != nil {
		return authmodel.User{}, fmt.Errorf("UserService.ElevateToAdmin: %w", err)
	}
	if hasAdmin {
		return authmodel.User{}, model.ErrAdminExists
	}
	user, err := s.repo.Update(ctx, userID, "", "", authmodel.RoleAdmin)
	if err != nil {
		return authmodel.User{}, fmt.Errorf("UserService.ElevateToAdmin: %w", err)
	}
	return user, nil
}
