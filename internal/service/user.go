package service

import (
	"context"

	"starter/internal/model"
	"starter/internal/repository"

	"github.com/google/uuid"
)

// UserService provides CRUD operations for user management.
type UserService struct {
	repo repository.UserRepository
}

// List returns a paginated slice of users. perPage is capped at 100.
func (s *UserService) List(ctx context.Context, page, perPage int) ([]model.User, error) {
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage
	return s.repo.List(ctx, int32(perPage), int32(offset))
}

// GetByID retrieves a single user by ID. Returns model.ErrUserNotFound when missing.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	return s.repo.GetByID(ctx, id)
}

// Update applies non-empty fields from input to the user. Empty strings are
// treated as "no change" by the underlying COALESCE SQL logic.
func (s *UserService) Update(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (model.User, error) {
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
