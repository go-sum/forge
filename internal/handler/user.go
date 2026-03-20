package handler

import (
	"net/http"

	"starter/internal/model"
	"starter/internal/view/page"
	"starter/internal/view/partial/userpartial"
	pkgform "starter/pkg/components/patterns/form"
	"starter/pkg/components/patterns/flash"
	"starter/pkg/components/patterns/pager"
	"starter/pkg/ctxkeys"
	"starter/pkg/render"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// UserList renders the paginated user management table.
func (h *Handler) UserList(c *echo.Context) error {
	ctx := c.Request().Context()
	userID, _ := c.Get(string(ctxkeys.UserID)).(string)

	pg := pager.New(c.Request(), 20)

	total, _ := h.services.User.Count(ctx)
	pg.SetTotal(int(total))

	users, _ := h.services.User.List(ctx, pg.Page, pg.PerPage)

	flashMsgs, _ := flash.GetAll(c.Request(), c.Response())

	return render.Component(c, page.UserListPage(page.UserListProps{
		Users:           users,
		Pager:           pg,
		CSRFToken:       h.csrfToken(c),
		Flash:           flashMsgs,
		IsAuthenticated: userID != "",
	}))
}

// UserEditForm renders the inline edit form for a single user row (HTMX swap).
func (h *Handler) UserEditForm(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	user, err := h.services.User.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return render.Fragment(c, userpartial.UserEditForm(userpartial.UserFormProps{
		User:      user,
		CSRFToken: h.csrfToken(c),
	}))
}

// UserRow renders a single read-only user table row (HTMX swap after save/cancel).
func (h *Handler) UserRow(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	user, err := h.services.User.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return render.Fragment(c, userpartial.UserRow(userpartial.UserRowProps{
		User:      user,
		CSRFToken: h.csrfToken(c),
	}))
}

// UserUpdate processes the inline edit form and returns the updated row.
func (h *Handler) UserUpdate(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	var input model.UpdateUserInput
	sub := pkgform.New(h.validator.Validate())
	sub.Submit(c, &input)

	if !sub.IsValid() {
		user, _ := h.services.User.GetByID(c.Request().Context(), id)
		errs := sub.GetErrors()
		// Remap struct field names (capitalized) to form field names (snake_case).
		formErrors := map[string][]string{
			"email":        errs["Email"],
			"display_name": errs["DisplayName"],
			"role":         errs["Role"],
		}
		return render.FragmentWithStatus(c, http.StatusUnprocessableEntity, userpartial.UserEditForm(userpartial.UserFormProps{
			User:      user,
			CSRFToken: h.csrfToken(c),
			Errors:    formErrors,
		}))
	}

	user, err := h.services.User.Update(c.Request().Context(), id, input)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return render.Fragment(c, userpartial.UserRow(userpartial.UserRowProps{
		User:      user,
		CSRFToken: h.csrfToken(c),
	}))
}

// UserDelete removes a user and returns 200 so HTMX can remove the row.
func (h *Handler) UserDelete(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	if err := h.services.User.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
