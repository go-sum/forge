// Package service implements domain business logic. Services orchestrate
// operations across repositories and apply business rules. They know nothing
// about HTTP — that belongs to the handler layer.
package service

import (
	"starter/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Services is the composition root for all domain services.
type Services struct {
	Auth *AuthService
	User *UserService
}

// NewServices constructs all domain services from the shared infrastructure.
// Sessions are not injected here — they belong to the transport layer.
func NewServices(repos *repository.Repositories, pool *pgxpool.Pool) *Services {
	return &Services{
		Auth: NewAuthService(repos.User, repos.Password, repos, pool),
		User: NewUserService(repos.User),
	}
}
