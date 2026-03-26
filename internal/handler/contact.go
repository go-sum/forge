package handler

import (
	"net/http"

	"github.com/go-sum/componentry/patterns/form"
	render "github.com/go-sum/componentry/render/echo"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/page"
	"github.com/go-sum/forge/internal/view/partial/contactpartial"
	"github.com/go-sum/server/apperr"

	"github.com/labstack/echo/v5"
)

// ContactForm renders the contact-us page.
func (h *Handler) ContactForm(c *echo.Context) error {
	req := h.request(c)
	data := model.ContactFormData{}
	return view.Render(c, req, page.ContactPage(req, data), contactpartial.ContactForm(req, data))
}

// ContactSubmit processes the contact form, sends the emails, and returns
// the form partial in either a success or validation-error state.
func (h *Handler) ContactSubmit(c *echo.Context) error {
	req := h.request(c)

	var input model.ContactInput
	sub := form.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		return render.FragmentWithStatus(c, http.StatusUnprocessableEntity, contactpartial.ContactForm(req, model.ContactFormData{
			Values: input,
			Errors: sub.GetErrors(),
		}))
	}

	if err := h.services.Contact.Submit(c.Request().Context(), input); err != nil {
		return apperr.Unavailable("Unable to send your message right now. Please try again later.", err)
	}

	return render.Fragment(c, contactpartial.ContactForm(req, model.ContactFormData{Sent: true}))
}
