package service

import (
	"context"
	"errors"
	"testing"
	"time"

	authmodel "github.com/go-sum/auth/model"
	"github.com/go-sum/forge/internal/model"

	"github.com/google/uuid"
)

var serviceTestUser = authmodel.User{
	ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	Email:       "ada@example.com",
	DisplayName: "Ada Lovelace",
	Role:        "admin",
	CreatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
	UpdatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
}

type fakeUserRepo struct {
	listFn     func(context.Context, int32, int32) ([]authmodel.User, error)
	getByID    func(context.Context, uuid.UUID) (authmodel.User, error)
	getByEmail func(context.Context, string) (authmodel.User, error)
	updateFn   func(context.Context, uuid.UUID, string, string, string) (authmodel.User, error)
	deleteFn   func(context.Context, uuid.UUID) error
	countFn    func(context.Context) (int64, error)
	hasAdminFn func(context.Context) (bool, error)
}

func (r fakeUserRepo) GetByID(ctx context.Context, id uuid.UUID) (authmodel.User, error) {
	if r.getByID != nil {
		return r.getByID(ctx, id)
	}
	return authmodel.User{}, errors.New("unexpected GetByID call")
}

func (r fakeUserRepo) GetByEmail(ctx context.Context, email string) (authmodel.User, error) {
	if r.getByEmail != nil {
		return r.getByEmail(ctx, email)
	}
	return authmodel.User{}, errors.New("unexpected GetByEmail call")
}

func (r fakeUserRepo) List(ctx context.Context, limit, offset int32) ([]authmodel.User, error) {
	if r.listFn != nil {
		return r.listFn(ctx, limit, offset)
	}
	return nil, errors.New("unexpected List call")
}

func (r fakeUserRepo) Update(ctx context.Context, id uuid.UUID, email, displayName, role string) (authmodel.User, error) {
	if r.updateFn != nil {
		return r.updateFn(ctx, id, email, displayName, role)
	}
	return authmodel.User{}, errors.New("unexpected Update call")
}

func (r fakeUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if r.deleteFn != nil {
		return r.deleteFn(ctx, id)
	}
	return errors.New("unexpected Delete call")
}

func (r fakeUserRepo) Count(ctx context.Context) (int64, error) {
	if r.countFn != nil {
		return r.countFn(ctx)
	}
	return 0, errors.New("unexpected Count call")
}

func (r fakeUserRepo) HasAdmin(ctx context.Context) (bool, error) {
	if r.hasAdminFn != nil {
		return r.hasAdminFn(ctx)
	}
	return false, errors.New("unexpected HasAdmin call")
}

func TestUserServiceListCapsPerPageAndComputesOffset(t *testing.T) {
	svc := NewUserService(fakeUserRepo{
		listFn: func(_ context.Context, limit, offset int32) ([]authmodel.User, error) {
			if limit != 100 || offset != 200 {
				t.Fatalf("limit=%d offset=%d", limit, offset)
			}
			return []authmodel.User{serviceTestUser}, nil
		},
	})

	users, err := svc.List(context.Background(), 3, 250)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(users) != 1 || users[0].ID != serviceTestUser.ID {
		t.Fatalf("users = %#v", users)
	}
}

func TestUserServiceDelegatesCRUDOperations(t *testing.T) {
	svc := NewUserService(fakeUserRepo{
		getByID: func(context.Context, uuid.UUID) (authmodel.User, error) { return serviceTestUser, nil },
		updateFn: func(_ context.Context, id uuid.UUID, email, displayName, role string) (authmodel.User, error) {
			if id != serviceTestUser.ID || email != "" || displayName != "" || role != "" {
				t.Fatalf("id=%s email=%q displayName=%q role=%q", id, email, displayName, role)
			}
			return serviceTestUser, nil
		},
		deleteFn: func(context.Context, uuid.UUID) error { return nil },
		countFn:  func(context.Context) (int64, error) { return 42, nil },
	})

	user, err := svc.GetByID(context.Background(), serviceTestUser.ID)
	if err != nil || user.ID != serviceTestUser.ID {
		t.Fatalf("GetByID() user=%#v err=%v", user, err)
	}

	user, err = svc.Update(context.Background(), serviceTestUser.ID, authmodel.UpdateUserInput{})
	if err != nil || user.ID != serviceTestUser.ID {
		t.Fatalf("Update() user=%#v err=%v", user, err)
	}

	if err := svc.Delete(context.Background(), serviceTestUser.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	count, err := svc.Count(context.Background())
	if err != nil || count != 42 {
		t.Fatalf("Count() count=%d err=%v", count, err)
	}
}

func TestUserServiceHasAdminDelegatesToRepo(t *testing.T) {
	tests := []struct {
		name    string
		repoVal bool
		repoErr error
		wantVal bool
		wantErr bool
	}{
		{
			name:    "repo returns true",
			repoVal: true,
			wantVal: true,
		},
		{
			name:    "repo returns false",
			repoVal: false,
			wantVal: false,
		},
		{
			name:    "repo returns error",
			repoErr: errors.New("db down"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewUserService(fakeUserRepo{
				hasAdminFn: func(context.Context) (bool, error) {
					return tt.repoVal, tt.repoErr
				},
			})

			got, err := svc.HasAdmin(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("HasAdmin() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("HasAdmin() error = %v", err)
			}
			if got != tt.wantVal {
				t.Fatalf("HasAdmin() = %v, want %v", got, tt.wantVal)
			}
		})
	}
}

func TestUserServiceElevateToAdmin(t *testing.T) {
	adminUser := serviceTestUser
	adminUser.Role = authmodel.RoleAdmin

	tests := []struct {
		name       string
		hasAdminFn func(context.Context) (bool, error)
		updateFn   func(context.Context, uuid.UUID, string, string, string) (authmodel.User, error)
		wantUser   authmodel.User
		wantErr    error
	}{
		{
			name: "happy path promotes user to admin",
			hasAdminFn: func(context.Context) (bool, error) {
				return false, nil
			},
			updateFn: func(_ context.Context, id uuid.UUID, email, displayName, role string) (authmodel.User, error) {
				if id != serviceTestUser.ID {
					t.Fatalf("Update id = %s, want %s", id, serviceTestUser.ID)
				}
				if email != "" {
					t.Fatalf("Update email = %q, want empty", email)
				}
				if displayName != "" {
					t.Fatalf("Update displayName = %q, want empty", displayName)
				}
				if role != authmodel.RoleAdmin {
					t.Fatalf("Update role = %q, want %q", role, authmodel.RoleAdmin)
				}
				return adminUser, nil
			},
			wantUser: adminUser,
		},
		{
			name: "admin already exists",
			hasAdminFn: func(context.Context) (bool, error) {
				return true, nil
			},
			wantErr: model.ErrAdminExists,
		},
		{
			name: "HasAdmin repo error propagated",
			hasAdminFn: func(context.Context) (bool, error) {
				return false, errors.New("db down")
			},
			wantErr: errors.New("db down"),
		},
		{
			name: "Update returns user not found",
			hasAdminFn: func(context.Context) (bool, error) {
				return false, nil
			},
			updateFn: func(context.Context, uuid.UUID, string, string, string) (authmodel.User, error) {
				return authmodel.User{}, authmodel.ErrUserNotFound
			},
			wantErr: authmodel.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewUserService(fakeUserRepo{
				hasAdminFn: tt.hasAdminFn,
				updateFn:   tt.updateFn,
			})

			user, err := svc.ElevateToAdmin(context.Background(), serviceTestUser.ID)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("ElevateToAdmin() expected error, got nil")
				}
				if errors.Is(tt.wantErr, model.ErrAdminExists) && !errors.Is(err, model.ErrAdminExists) {
					t.Fatalf("ElevateToAdmin() error = %v, want %v", err, model.ErrAdminExists)
				}
				if errors.Is(tt.wantErr, authmodel.ErrUserNotFound) && !errors.Is(err, authmodel.ErrUserNotFound) {
					t.Fatalf("ElevateToAdmin() error = %v, want %v", err, authmodel.ErrUserNotFound)
				}
				return
			}
			if err != nil {
				t.Fatalf("ElevateToAdmin() error = %v", err)
			}
			if user.ID != tt.wantUser.ID {
				t.Fatalf("ElevateToAdmin() user.ID = %s, want %s", user.ID, tt.wantUser.ID)
			}
			if user.Role != tt.wantUser.Role {
				t.Fatalf("ElevateToAdmin() user.Role = %q, want %q", user.Role, tt.wantUser.Role)
			}
		})
	}
}
