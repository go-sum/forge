// Package service implements domain business logic. Services orchestrate
// operations across repositories and apply business rules. They know nothing
// about HTTP — that belongs to the handler layer.
package service

import (
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/repository"
	"github.com/go-sum/send"
)

// Services is the composition root for all domain services.
type Services struct {
	User    *UserService
	Contact *ContactService
}

// NewServices constructs all domain services from the shared infrastructure.
// Auth is now handled by the auth module's AuthService, wired in container.go.
func NewServices(repos *repository.Repositories, sender send.Sender, sendCfg *config.SendConfig) *Services {
	return &Services{
		User:    NewUserService(repos.User),
		Contact: NewContactService(sender, sendCfg),
	}
}
