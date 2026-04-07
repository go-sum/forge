// Package service implements domain business logic. Services orchestrate
// operations across repositories and apply business rules. They know nothing
// about HTTP — that belongs to the handler layer.
package service

import (
	authrepo "github.com/go-sum/auth/repository"
	"github.com/go-sum/queue"
)

// Services is the composition root for all domain services.
type Services struct {
	User    *UserService
	Contact *ContactService
}

// NewServices constructs all domain services from the shared infrastructure.
// Auth is now handled by the auth module's AuthService, wired in container.go.
func NewServices(adminStore authrepo.AdminStore, q *queue.Client, contactCfg ContactConfig) *Services {
	return &Services{
		User:    NewUserService(adminStore),
		Contact: NewContactService(q, contactCfg),
	}
}
