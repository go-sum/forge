package repository

import (
	"context"
	"errors"

	db "starter/db/schema"
	"starter/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type passwordRepository struct{ q *db.Queries }

func newPasswordRepository(dbtx db.DBTX) *passwordRepository {
	return &passwordRepository{q: db.New(dbtx)}
}

func toPasswordModel(p db.Password) model.Password {
	return model.Password{
		ID:        p.ID,
		UserID:    p.UserID,
		Hash:      p.Hash,
		CreatedAt: p.CreatedAt,
	}
}

func (r *passwordRepository) Create(ctx context.Context, userID uuid.UUID, hash string) (model.Password, error) {
	p, err := r.q.CreatePassword(ctx, db.CreatePasswordParams{UserID: userID, Hash: hash})
	if err != nil {
		return model.Password{}, err
	}
	return toPasswordModel(p), nil
}

func (r *passwordRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, hash string) (model.Password, error) {
	p, err := db.New(tx).CreatePassword(ctx, db.CreatePasswordParams{UserID: userID, Hash: hash})
	if err != nil {
		return model.Password{}, err
	}
	return toPasswordModel(p), nil
}

func (r *passwordRepository) GetCurrentByUserID(ctx context.Context, userID uuid.UUID) (model.Password, error) {
	p, err := r.q.GetCurrentPasswordByUserID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Password{}, model.ErrUserNotFound
	}
	if err != nil {
		return model.Password{}, err
	}
	return toPasswordModel(p), nil
}

func (r *passwordRepository) GetCurrentByEmail(ctx context.Context, email string) (model.Password, error) {
	p, err := r.q.GetCurrentPasswordByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		// No user enumeration: treat missing record as invalid credentials.
		return model.Password{}, model.ErrInvalidCredentials
	}
	if err != nil {
		return model.Password{}, err
	}
	return toPasswordModel(p), nil
}

func (r *passwordRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.Password, error) {
	rows, err := r.q.ListPasswordsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	passwords := make([]model.Password, len(rows))
	for i, p := range rows {
		passwords[i] = toPasswordModel(p)
	}
	return passwords, nil
}
