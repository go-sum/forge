// Example: safe to delete as a unit.
package repository

import (
	"context"
	"errors"

	db "github.com/go-sum/forge/db/schema"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/server/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type userRepository struct{ q *db.Queries }

func newUserRepository(dbtx db.DBTX) *userRepository {
	return &userRepository{q: db.New(dbtx)}
}

func toUserModel(u db.User) model.User {
	return model.User{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		Verified:    u.Verified,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func (r *userRepository) Create(ctx context.Context, email, displayName, role string, verified bool) (model.User, error) {
	u, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email:       email,
		DisplayName: displayName,
		Role:        role,
		Verified:    verified,
	})
	if err != nil {
		return model.User{}, mapUserErr(err)
	}
	return toUserModel(u), nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return toUserModel(u), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return toUserModel(u), nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int32) ([]model.User, error) {
	rows, err := r.q.ListUsers(ctx, db.ListUsersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	users := make([]model.User, len(rows))
	for i, u := range rows {
		users[i] = toUserModel(u)
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, id uuid.UUID, email, displayName, role string) (model.User, error) {
	u, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:          id,
		Email:       email,
		DisplayName: displayName,
		Role:        role,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, mapUserErr(err)
	}
	return toUserModel(u), nil
}

func (r *userRepository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (model.User, error) {
	u, err := r.q.UpdateUserEmail(ctx, db.UpdateUserEmailParams{
		ID:    id,
		Email: email,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, mapUserErr(err)
	}
	return toUserModel(u), nil
}

// Delete removes a user. If no row matches id, the operation succeeds silently —
// DELETE is idempotent per RFC 9110 §9.3.5.
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteUser(ctx, id)
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountUsers(ctx)
}

// mapUserErr translates Postgres unique constraint violations to domain errors.
func mapUserErr(err error) error {
	if database.IsUniqueViolation(err) {
		return model.ErrEmailTaken
	}
	return err
}
