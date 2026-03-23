package adapters

import (
	"context"
	"errors"

	authmodel "github.com/go-sum/auth/model"
	authrepo "github.com/go-sum/auth/repository"
	db "github.com/go-sum/forge/db/schema"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/repository"
	"github.com/go-sum/server/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- UserReader adapter ---

// authUserReader adapts forge's UserRepository to auth's UserReader port.
type authUserReader struct{ repo repository.UserRepository }

// NewAuthUserReader wraps forge's user repository to satisfy auth/repository.UserReader.
func NewAuthUserReader(repo repository.UserRepository) authrepo.UserReader {
	return &authUserReader{repo: repo}
}

func (a *authUserReader) GetByID(ctx context.Context, id uuid.UUID) (authmodel.User, error) {
	u, err := a.repo.GetByID(ctx, id)
	return toAuthUser(u, err)
}

func (a *authUserReader) GetByEmail(ctx context.Context, email string) (authmodel.User, error) {
	u, err := a.repo.GetByEmail(ctx, email)
	return toAuthUser(u, err)
}

// --- PasswordStore adapter ---

// authPasswordStore adapts forge's PasswordRepository to auth's PasswordStore port.
type authPasswordStore struct{ repo repository.PasswordRepository }

// NewAuthPasswordStore wraps forge's password repository to satisfy auth/repository.PasswordStore.
func NewAuthPasswordStore(repo repository.PasswordRepository) authrepo.PasswordStore {
	return &authPasswordStore{repo: repo}
}

func (a *authPasswordStore) Create(ctx context.Context, userID uuid.UUID, hash string) (authmodel.Password, error) {
	p, err := a.repo.Create(ctx, userID, hash)
	if err != nil {
		return authmodel.Password{}, err
	}
	return authmodel.Password{
		ID:        p.ID,
		UserID:    p.UserID,
		Hash:      p.Hash,
		CreatedAt: p.CreatedAt,
	}, nil
}

func (a *authPasswordStore) GetCurrentByEmail(ctx context.Context, email string) (authmodel.Password, error) {
	p, err := a.repo.GetCurrentByEmail(ctx, email)
	if errors.Is(err, model.ErrInvalidCredentials) {
		return authmodel.Password{}, authmodel.ErrInvalidCredentials
	}
	if errors.Is(err, model.ErrUserNotFound) {
		return authmodel.Password{}, authmodel.ErrUserNotFound
	}
	if err != nil {
		return authmodel.Password{}, err
	}
	return authmodel.Password{
		ID:        p.ID,
		UserID:    p.UserID,
		Hash:      p.Hash,
		CreatedAt: p.CreatedAt,
	}, nil
}

func (a *authPasswordStore) GetCurrentByUserID(ctx context.Context, userID uuid.UUID) (authmodel.Password, error) {
	p, err := a.repo.GetCurrentByUserID(ctx, userID)
	if errors.Is(err, model.ErrUserNotFound) {
		return authmodel.Password{}, authmodel.ErrUserNotFound
	}
	if err != nil {
		return authmodel.Password{}, err
	}
	return authmodel.Password{
		ID:        p.ID,
		UserID:    p.UserID,
		Hash:      p.Hash,
		CreatedAt: p.CreatedAt,
	}, nil
}

// --- TxFactory ---

// authTxFactory implements auth/service.txFactory using forge's db package.
type authTxFactory struct{ pool *pgxpool.Pool }

// NewAuthTxFactory returns a txFactory that creates auth-compatible tx-scoped repositories.
func NewAuthTxFactory(pool *pgxpool.Pool) interface {
	WithTx(pgx.Tx) authrepo.TxRepos
} {
	return &authTxFactory{pool: pool}
}

func (f *authTxFactory) WithTx(tx pgx.Tx) authrepo.TxRepos {
	q := db.New(tx)
	return authrepo.TxRepos{
		User:     &txAuthUserWriter{q: q},
		Password: &txAuthPasswordStore{q: q},
	}
}

// txAuthUserWriter implements auth/repository.UserWriter using sqlc queries within a tx.
type txAuthUserWriter struct{ q *db.Queries }

func (r *txAuthUserWriter) Create(ctx context.Context, email, displayName, role string) (authmodel.User, error) {
	u, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email:       email,
		DisplayName: displayName,
		Role:        role,
	})
	if err != nil {
		return authmodel.User{}, mapAuthUserErr(err)
	}
	return authmodel.User{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (r *txAuthUserWriter) GetByID(ctx context.Context, id uuid.UUID) (authmodel.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return authmodel.User{}, authmodel.ErrUserNotFound
	}
	if err != nil {
		return authmodel.User{}, err
	}
	return authmodel.User{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (r *txAuthUserWriter) GetByEmail(ctx context.Context, email string) (authmodel.User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return authmodel.User{}, authmodel.ErrUserNotFound
	}
	if err != nil {
		return authmodel.User{}, err
	}
	return authmodel.User{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

// txAuthPasswordStore implements auth/repository.PasswordStore using sqlc queries within a tx.
type txAuthPasswordStore struct{ q *db.Queries }

func (r *txAuthPasswordStore) Create(ctx context.Context, userID uuid.UUID, hash string) (authmodel.Password, error) {
	p, err := r.q.CreatePassword(ctx, db.CreatePasswordParams{UserID: userID, Hash: hash})
	if err != nil {
		return authmodel.Password{}, err
	}
	return authmodel.Password{
		ID:        p.ID,
		UserID:    p.UserID,
		Hash:      p.Hash,
		CreatedAt: p.CreatedAt,
	}, nil
}

func (r *txAuthPasswordStore) GetCurrentByEmail(ctx context.Context, email string) (authmodel.Password, error) {
	p, err := r.q.GetCurrentPasswordByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return authmodel.Password{}, authmodel.ErrInvalidCredentials
	}
	if err != nil {
		return authmodel.Password{}, err
	}
	return authmodel.Password{
		ID:        p.ID,
		UserID:    p.UserID,
		Hash:      p.Hash,
		CreatedAt: p.CreatedAt,
	}, nil
}

func (r *txAuthPasswordStore) GetCurrentByUserID(ctx context.Context, userID uuid.UUID) (authmodel.Password, error) {
	p, err := r.q.GetCurrentPasswordByUserID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return authmodel.Password{}, authmodel.ErrUserNotFound
	}
	if err != nil {
		return authmodel.Password{}, err
	}
	return authmodel.Password{
		ID:        p.ID,
		UserID:    p.UserID,
		Hash:      p.Hash,
		CreatedAt: p.CreatedAt,
	}, nil
}

// --- helpers ---

func toAuthUser(u model.User, err error) (authmodel.User, error) {
	if errors.Is(err, model.ErrUserNotFound) {
		return authmodel.User{}, authmodel.ErrUserNotFound
	}
	if err != nil {
		return authmodel.User{}, err
	}
	return authmodel.User{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func mapAuthUserErr(err error) error {
	if database.IsUniqueViolation(err) {
		return authmodel.ErrEmailTaken
	}
	return err
}
