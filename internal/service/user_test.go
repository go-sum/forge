package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-sum/forge/internal/model"

	"github.com/google/uuid"
)

var serviceTestUser = model.User{
	ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	Email:       "ada@example.com",
	DisplayName: "Ada Lovelace",
	Role:        "admin",
	CreatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
	UpdatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
}

type fakeUserRepo struct {
	listFn   func(context.Context, int32, int32) ([]model.User, error)
	getByID  func(context.Context, uuid.UUID) (model.User, error)
	updateFn func(context.Context, uuid.UUID, string, string, string) (model.User, error)
	deleteFn func(context.Context, uuid.UUID) error
	countFn  func(context.Context) (int64, error)
}

func (fakeUserRepo) Create(context.Context, string, string, string, bool) (model.User, error) {
	return model.User{}, errors.New("unexpected Create call")
}

func (r fakeUserRepo) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	if r.getByID != nil {
		return r.getByID(ctx, id)
	}
	return model.User{}, errors.New("unexpected GetByID call")
}

func (fakeUserRepo) GetByEmail(context.Context, string) (model.User, error) {
	return model.User{}, errors.New("unexpected GetByEmail call")
}

func (r fakeUserRepo) List(ctx context.Context, limit, offset int32) ([]model.User, error) {
	if r.listFn != nil {
		return r.listFn(ctx, limit, offset)
	}
	return nil, errors.New("unexpected List call")
}

func (r fakeUserRepo) Update(ctx context.Context, id uuid.UUID, email, displayName, role string) (model.User, error) {
	if r.updateFn != nil {
		return r.updateFn(ctx, id, email, displayName, role)
	}
	return model.User{}, errors.New("unexpected Update call")
}

func (fakeUserRepo) UpdateEmail(context.Context, uuid.UUID, string) (model.User, error) {
	return model.User{}, errors.New("unexpected UpdateEmail call")
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

func TestUserServiceListCapsPerPageAndComputesOffset(t *testing.T) {
	svc := NewUserService(fakeUserRepo{
		listFn: func(_ context.Context, limit, offset int32) ([]model.User, error) {
			if limit != 100 || offset != 200 {
				t.Fatalf("limit=%d offset=%d", limit, offset)
			}
			return []model.User{serviceTestUser}, nil
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
		getByID: func(context.Context, uuid.UUID) (model.User, error) { return serviceTestUser, nil },
		updateFn: func(_ context.Context, id uuid.UUID, email, displayName, role string) (model.User, error) {
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

	user, err = svc.Update(context.Background(), serviceTestUser.ID, model.UpdateUserInput{})
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
