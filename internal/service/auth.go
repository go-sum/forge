package service

import (
	"context"
	"errors"
	"fmt"

	"starter/internal/model"
	"starter/internal/repository"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// txFactory creates transaction-scoped repositories.
// *repository.Repositories satisfies this interface via its WithTx method.
type txFactory interface {
	WithTx(pgx.Tx) *repository.Repositories
}

type txBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

// AuthService handles user registration and login.
type AuthService struct {
	users     repository.UserRepository
	passwords repository.PasswordRepository
	factory   txFactory
	pool      txBeginner
}

// NewAuthService constructs an AuthService from its repository and transaction dependencies.
func NewAuthService(
	users repository.UserRepository,
	passwords repository.PasswordRepository,
	factory txFactory,
	pool txBeginner,
) *AuthService {
	return &AuthService{
		users:     users,
		passwords: passwords,
		factory:   factory,
		pool:      pool,
	}
}

// Register creates a new user and their initial password in a single transaction.
// Returns model.ErrEmailTaken when the email is already in use.
func (s *AuthService) Register(ctx context.Context, input model.CreateUserInput) (model.User, error) {
	role := input.Role
	if role == "" {
		role = model.RoleUser
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, fmt.Errorf("hash password: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return model.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	txRepos := s.factory.WithTx(tx)

	user, err := txRepos.User.Create(ctx, input.Email, input.DisplayName, role)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			return model.User{}, err
		}
		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	if _, err := txRepos.Password.Create(ctx, user.ID, string(hash)); err != nil {
		return model.User{}, fmt.Errorf("create password: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.User{}, fmt.Errorf("commit tx: %w", err)
	}

	return user, nil
}

// Login authenticates a user by email and password.
// Always returns model.ErrInvalidCredentials on auth failure to prevent user enumeration.
func (s *AuthService) Login(ctx context.Context, input model.LoginInput) (model.User, error) {
	pwd, err := s.passwords.GetCurrentByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, model.ErrInvalidCredentials) || errors.Is(err, model.ErrUserNotFound) {
			return model.User{}, model.ErrInvalidCredentials
		}
		return model.User{}, fmt.Errorf("get password: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pwd.Hash), []byte(input.Password)); err != nil {
		return model.User{}, model.ErrInvalidCredentials
	}

	user, err := s.users.GetByID(ctx, pwd.UserID)
	if err != nil {
		return model.User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}
