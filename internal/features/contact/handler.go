package contact

import (
	"context"
	"net/http"

	"github.com/go-sum/componentry/patterns/form"
	render "github.com/go-sum/componentry/render/echo"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/page"
	"github.com/go-sum/forge/internal/view/partial/contactpartial"
	"github.com/go-sum/server/apperr"

	"github.com/labstack/echo/v5"
)

type submitter interface {
	Submit(ctx context.Context, input model.ContactInput) error
}

type Handler struct {
	cfg       *config.Config
	service   submitter
	validator form.StructValidator
}

func NewHandler(cfg *config.Config, service submitter, validator form.StructValidator) *Handler {
	return &Handler{cfg: cfg, service: service, validator: validator}
}

func (h *Handler) Form(c *echo.Context) error {
	req := view.NewRequest(c, h.cfg)
	data := model.ContactFormData{}
	return view.Render(c, req, page.ContactPage(req, data), contactpartial.ContactForm(req, data))
}

func (h *Handler) Submit(c *echo.Context) error {
	req := view.NewRequest(c, h.cfg)

	var input model.ContactInput
	sub := form.New(h.validator)
	sub.Submit(c, &input)

	if !sub.IsValid() {
		return render.FragmentWithStatus(c, http.StatusUnprocessableEntity, contactpartial.ContactForm(req, model.ContactFormData{
			Values: input,
			Errors: sub.GetErrors(),
		}))
	}

	if err := h.service.Submit(c.Request().Context(), input); err != nil {
		return apperr.Unavailable("Unable to send your message right now. Please try again later.", err)
	}

	return render.Fragment(c, contactpartial.ContactForm(req, model.ContactFormData{Sent: true}))
}
