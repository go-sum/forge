package repository

import (
	"context"
	"errors"

	db "starter/db/schema"
	"starter/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func (r *userRepository) Create(ctx context.Context, email, displayName, role string) (model.User, error) {
	u, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email:       email,
		DisplayName: displayName,
		Role:        role,
	})
	if err != nil {
		return model.User{}, mapUserErr(err)
	}
	return toUserModel(u), nil
}

func (r *userRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, email, displayName, role string) (model.User, error) {
	u, err := db.New(tx).CreateUser(ctx, db.CreateUserParams{
		Email:       email,
		DisplayName: displayName,
		Role:        role,
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

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteUser(ctx, id)
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountUsers(ctx)
}

// mapUserErr translates Postgres unique constraint violations to domain errors.
func mapUserErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return model.ErrEmailTaken
	}
	return err
}
