package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
)

type fakeAdminStore struct {
	getByIDFn func(context.Context, uuid.UUID) (model.User, error)
	updateFn  func(context.Context, uuid.UUID, string, string, string) (model.User, error)
	deleteFn  func(context.Context, uuid.UUID) error
	countFn   func(context.Context) (int64, error)
	listFn    func(context.Context, int32, int32) ([]model.User, error)
	hasAdmin  func(context.Context) (bool, error)
}

func (f fakeAdminStore) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return model.User{}, errors.New("unexpected GetByID call")
}

func (f fakeAdminStore) GetByEmail(context.Context, string) (model.User, error) {
	return model.User{}, errors.New("unexpected GetByEmail call")
}

func (f fakeAdminStore) List(ctx context.Context, limit, offset int32) ([]model.User, error) {
	if f.listFn != nil {
		return f.listFn(ctx, limit, offset)
	}
	return nil, errors.New("unexpected List call")
}

func (f fakeAdminStore) Update(ctx context.Context, id uuid.UUID, email, displayName, role string) (model.User, error) {
	if f.updateFn != nil {
		return f.updateFn(ctx, id, email, displayName, role)
	}
	return model.User{}, errors.New("unexpected Update call")
}

func (f fakeAdminStore) Delete(ctx context.Context, id uuid.UUID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return errors.New("unexpected Delete call")
}

func (f fakeAdminStore) Count(ctx context.Context) (int64, error) {
	if f.countFn != nil {
		return f.countFn(ctx)
	}
	return 0, errors.New("unexpected Count call")
}

func (f fakeAdminStore) HasAdmin(ctx context.Context) (bool, error) {
	if f.hasAdmin != nil {
		return f.hasAdmin(ctx)
	}
	return false, errors.New("unexpected HasAdmin call")
}

func TestAdminServiceListUsersPassesThroughPerPage(t *testing.T) {
	svc := NewAdminService(fakeAdminStore{
		listFn: func(_ context.Context, limit, offset int32) ([]model.User, error) {
			// Service passes caller's perPage through; capping is the handler's responsibility.
			if limit != 250 || offset != 500 {
				t.Fatalf("limit=%d offset=%d", limit, offset)
			}
			return []model.User{serviceTestUser}, nil
		},
	})

	users, err := svc.ListUsers(context.Background(), 3, 250)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 1 || users[0].ID != serviceTestUser.ID {
		t.Fatalf("users = %#v", users)
	}
}

func TestAdminServiceElevateToAdmin(t *testing.T) {
	tests := []struct {
		name     string
		hasAdmin func(context.Context) (bool, error)
		updateFn func(context.Context, uuid.UUID, string, string, string) (model.User, error)
		wantErr  error
	}{
		{
			name: "success",
			hasAdmin: func(context.Context) (bool, error) {
				return false, nil
			},
			updateFn: func(_ context.Context, id uuid.UUID, email, displayName, role string) (model.User, error) {
				if id != serviceTestUser.ID || email != "" || displayName != "" || role != model.RoleAdmin {
					t.Fatalf("id=%s email=%q displayName=%q role=%q", id, email, displayName, role)
				}
				user := serviceTestUser
				user.Role = model.RoleAdmin
				return user, nil
			},
		},
		{
			name: "admin exists",
			hasAdmin: func(context.Context) (bool, error) {
				return true, nil
			},
			wantErr: model.ErrAdminExists,
		},
		{
			name: "store error",
			hasAdmin: func(context.Context) (bool, error) {
				return false, errors.New("db down")
			},
			wantErr: errors.New("db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAdminService(fakeAdminStore{
				hasAdmin: tt.hasAdmin,
				updateFn: tt.updateFn,
			})
			_, err := svc.ElevateToAdmin(context.Background(), serviceTestUser.ID)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("ElevateToAdmin() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("ElevateToAdmin() error = nil")
			}
			if errors.Is(tt.wantErr, model.ErrAdminExists) && !errors.Is(err, model.ErrAdminExists) {
				t.Fatalf("error = %v, want %v", err, model.ErrAdminExists)
			}
			if tt.wantErr.Error() == "db down" && !strings.Contains(err.Error(), "db down") {
				t.Fatalf("error = %v, want contains %q", err, "db down")
			}
		})
	}
}
