package service

import (
	"context"
	"fmt"

	"github.com/go-sum/auth"
	"github.com/go-sum/auth/model"
	"github.com/go-sum/auth/repository"
	"github.com/google/uuid"
)

// AdminService implements auth.AdminService using the auth admin store contract.
type AdminService struct {
	users repository.AdminStore
}

var _ auth.AdminService = (*AdminService)(nil)

func NewAdminService(users repository.AdminStore) *AdminService {
	return &AdminService{users: users}
}

func (s *AdminService) CountUsers(ctx context.Context) (int64, error) {
	return s.users.Count(ctx)
}

func (s *AdminService) ListUsers(ctx context.Context, page, perPage int) ([]model.User, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage
	return s.users.List(ctx, int32(perPage), int32(offset))
}

func (s *AdminService) GetUserByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *AdminService) UpdateUser(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (model.User, error) {
	return s.users.Update(ctx, id, input.Email, input.DisplayName, input.Role)
}

func (s *AdminService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.users.Delete(ctx, id)
}

func (s *AdminService) HasAdmin(ctx context.Context) (bool, error) {
	return s.users.HasAdmin(ctx)
}

func (s *AdminService) ElevateToAdmin(ctx context.Context, userID uuid.UUID) (model.User, error) {
	hasAdmin, err := s.users.HasAdmin(ctx)
	if err != nil {
		return model.User{}, fmt.Errorf("AdminService.ElevateToAdmin: %w", err)
	}
	if hasAdmin {
		return model.User{}, model.ErrAdminExists
	}
	user, err := s.users.Update(ctx, userID, "", "", model.RoleAdmin)
	if err != nil {
		return model.User{}, fmt.Errorf("AdminService.ElevateToAdmin: %w", err)
	}
	return user, nil
}
