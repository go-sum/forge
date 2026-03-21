// Package handler contains the HTTP transport layer. Each handler method
// parses request data, delegates to a service, and renders a response.
package handler

import (
	"context"

	"starter/internal/model"
	"starter/internal/service"
	"starter/internal/view"
	"starter/pkg/auth"
	uilayout "starter/pkg/components/ui/layout"
	"starter/pkg/validate"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type authService interface {
	Login(context.Context, model.LoginInput) (model.User, error)
	Register(context.Context, model.CreateUserInput) (model.User, error)
}

type userService interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context, page, perPage int) ([]model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
	Update(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type handlerServices struct {
	Auth authService
	User userService
}

// Handler holds the transport layer's dependencies.
type Handler struct {
	services    handlerServices
	sessions    *auth.SessionManager
	validator   *validate.Validator
	checkHealth func(context.Context) error
	navConfig   uilayout.NavConfig
}

// New constructs a Handler with all required dependencies.
// checkHealth is a closure over the DB pool — the handler never holds raw
// infrastructure references directly.
func New(
	services *service.Services,
	sessions *auth.SessionManager,
	validator *validate.Validator,
	checkHealth func(context.Context) error,
	navConfig uilayout.NavConfig,
) *Handler {
	return &Handler{
		services: handlerServices{
			Auth: services.Auth,
			User: services.User,
		},
		sessions:    sessions,
		validator:   validator,
		checkHealth: checkHealth,
		navConfig:   navConfig,
	}
}

func (h *Handler) request(c *echo.Context) view.Request {
	return view.NewRequest(c, h.navConfig)
}
