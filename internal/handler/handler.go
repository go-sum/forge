// Package handler contains the HTTP transport layer. Each handler method
// parses request data, delegates to a service, and renders a response.
package handler

import (
	"context"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/service"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/server/validate"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type userService interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context, page, perPage int) ([]model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
	Update(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type handlerServices struct {
	User userService
}

// Handler holds the transport layer's dependencies.
type Handler struct {
	services    handlerServices
	validator   *validate.Validator
	checkHealth func(context.Context) error
	navConfig   config.NavConfig
}

// New constructs a Handler with all required dependencies.
// checkHealth is a closure over the DB pool — the handler never holds raw
// infrastructure references directly.
func New(
	services *service.Services,
	validator *validate.Validator,
	checkHealth func(context.Context) error,
	navConfig config.NavConfig,
) *Handler {
	return &Handler{
		services: handlerServices{
			User: services.User,
		},
		validator:   validator,
		checkHealth: checkHealth,
		navConfig:   navConfig,
	}
}

func (h *Handler) request(c *echo.Context) view.Request {
	return view.NewRequest(c, h.navConfig)
}
