// Package handler contains the HTTP transport layer. Each handler method
// parses request data, delegates to a service, and renders a response.
package handler

import (
	"context"

	"github.com/go-sum/componentry/patterns/form"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/service"
	"github.com/go-sum/forge/internal/view"

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

type contactService interface {
	Submit(ctx context.Context, input model.ContactInput) error
}

type handlerServices struct {
	User    userService
	Contact contactService
}

// Handler holds the transport layer's dependencies.
type Handler struct {
	cfg       *config.Config
	routes    func() echo.Routes // lazy accessor; evaluated at request time
	services  handlerServices
	validator form.StructValidator
}

// New constructs a Handler with all required dependencies.
// checkHealth is a closure over the DB pool — the handler never holds raw
// infrastructure references directly.
// routes is a lazy accessor for the registered Echo routes, used by handlers
// that need to resolve named routes to URLs (e.g. sitemap generation).
func New(
	cfg *config.Config,
	routes func() echo.Routes,
	services *service.Services,
	validator form.StructValidator,
) *Handler {
	return &Handler{
		cfg:    cfg,
		routes: routes,
		services: handlerServices{
			User:    services.User,
			Contact: services.Contact,
		},
		validator: validator,
	}
}

func (h *Handler) request(c *echo.Context) view.Request {
	return view.NewRequest(c, h.cfg)
}
