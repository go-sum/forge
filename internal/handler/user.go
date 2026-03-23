// Example: safe to delete as a unit (along with service/user.go, repository/user.go,
// view/page/users.go, view/partial/userpartial/, and the user routes in app/routes.go).
package handler

import (
	"errors"
	"net/http"

	pkgform "github.com/go-sum/componentry/patterns/form"
	"github.com/go-sum/componentry/patterns/pager"
	render "github.com/go-sum/componentry/render/echo"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/page"
	"github.com/go-sum/forge/internal/view/partial/userpartial"
	"github.com/go-sum/server/apperr"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// UserList renders the paginated user management table.
func (h *Handler) UserList(c *echo.Context) error {
	ctx := c.Request().Context()
	req := h.request(c)

	pg := pager.New(c.Request(), 20)

	total, err := h.services.User.Count(ctx)
	if err != nil {
		return apperr.Unavailable("Unable to load users right now.", err)
	}
	pg.SetTotal(int(total))

	users, err := h.services.User.List(ctx, pg.Page, pg.PerPage)
	if err != nil {
		return apperr.Unavailable("Unable to load users right now.", err)
	}

	data := page.UserListData{
		Users: users,
		Pager: pg,
	}
	return view.Render(c, req, page.UserListPage(req, data), page.UserListRegion(req, data))
}

// UserEditForm renders the inline edit form for a single user row (HTMX swap).
func (h *Handler) UserEditForm(c *echo.Context) error {
	req := h.request(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.BadRequest("The user ID in the URL is invalid.")
	}
	user, err := h.services.User.GetByID(c.Request().Context(), id)
	if err != nil {
		return resolveErr(err)
	}
	return render.Fragment(c, userpartial.UserEditForm(req, userpartial.UserFormData{User: user}))
}

// UserRow renders a single read-only user table row (HTMX swap after save/cancel).
func (h *Handler) UserRow(c *echo.Context) error {
	req := h.request(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.BadRequest("The user ID in the URL is invalid.")
	}
	user, err := h.services.User.GetByID(c.Request().Context(), id)
	if err != nil {
		return resolveErr(err)
	}
	return render.Fragment(c, userpartial.UserRow(req, userpartial.UserRowProps{
		User: user,
	}))
}

// UserUpdate processes the inline edit form and returns the updated row.
func (h *Handler) UserUpdate(c *echo.Context) error {
	req := h.request(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.BadRequest("The user ID in the URL is invalid.")
	}

	var input model.UpdateUserInput
	sub := pkgform.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		return render.FragmentWithStatus(c, http.StatusUnprocessableEntity, userpartial.UserEditForm(req, userpartial.UserFormData{
			User:   model.User{ID: id},
			Values: input,
			Errors: sub.GetErrors(),
		}))
	}

	user, err := h.services.User.Update(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			sub.SetFieldError("Email", "Email already in use.")
			return render.FragmentWithStatus(c, http.StatusConflict, userpartial.UserEditForm(req, userpartial.UserFormData{
				User:   model.User{ID: id},
				Values: input,
				Errors: sub.GetErrors(),
			}))
		}
		return resolveErr(err)
	}

	return render.Fragment(c, userpartial.UserRow(req, userpartial.UserRowProps{
		User: user,
	}))
}

// resolveErr maps forge domain errors to typed apperr responses.
// server/apperr.From is domain-agnostic, so domain mapping lives here.
func resolveErr(err error) *apperr.Error {
	if errors.Is(err, model.ErrUserNotFound) {
		return apperr.NotFound("The requested user does not exist.")
	}
	return resolveErr(err)
}

// UserDelete removes a user and returns 204 so HTMX can remove the row.
func (h *Handler) UserDelete(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.BadRequest("The user ID in the URL is invalid.")
	}
	if err := h.services.User.Delete(c.Request().Context(), id); err != nil {
		return resolveErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}
