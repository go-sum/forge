// Package service implements domain business logic. Services orchestrate
// operations across repositories and apply business rules. They know nothing
// about HTTP — that belongs to the handler layer.
package service

import (
	"starter/internal/repository"
	"starter/pkg/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Services is the composition root for all domain services.
type Services struct {
	Auth *AuthService
	User *UserService
}

// NewServices constructs all domain services from the shared infrastructure.
// Sessions are not injected here — they belong to the transport layer.
func NewServices(repos *repository.Repositories, pool *pgxpool.Pool, jwt auth.JWTConfig) *Services {
	return &Services{
		Auth: &AuthService{repos: repos, pool: pool, jwt: jwt},
		User: &UserService{repo: repos.User},
	}
}
