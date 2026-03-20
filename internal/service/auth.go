package service

import (
	"context"
	"errors"
	"fmt"

	"starter/internal/model"
	"starter/internal/repository"
	"starter/pkg/auth"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles user registration and login.
type AuthService struct {
	repos *repository.Repositories
	pool  *pgxpool.Pool
	jwt   auth.JWTConfig
}

// Register creates a new user and their initial password in a single transaction.
// Returns model.ErrEmailTaken when the email is already in use.
func (s *AuthService) Register(ctx context.Context, input model.CreateUserInput) (model.User, error) {
	role := input.Role
	if role == "" {
		role = "user"
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

	txRepos := s.repos.WithTx(tx)

	user, err := txRepos.User.CreateWithTx(ctx, tx, input.Email, input.DisplayName, role)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			return model.User{}, err
		}
		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	if _, err := txRepos.Password.CreateWithTx(ctx, tx, user.ID, string(hash)); err != nil {
		return model.User{}, fmt.Errorf("create password: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.User{}, fmt.Errorf("commit tx: %w", err)
	}

	return user, nil
}

// Login authenticates a user by email and password, returning the user and a signed JWT.
// Always returns model.ErrInvalidCredentials on auth failure to prevent user enumeration.
func (s *AuthService) Login(ctx context.Context, input model.LoginInput) (model.User, string, error) {
	pwd, err := s.repos.Password.GetCurrentByEmail(ctx, input.Email)
	if err != nil {
		return model.User{}, "", model.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pwd.Hash), []byte(input.Password)); err != nil {
		return model.User{}, "", model.ErrInvalidCredentials
	}

	user, err := s.repos.User.GetByEmail(ctx, input.Email)
	if err != nil {
		return model.User{}, "", fmt.Errorf("get user: %w", err)
	}

	token, err := auth.GenerateToken(s.jwt, user.ID, user.Email, user.Role)
	if err != nil {
		return model.User{}, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}
