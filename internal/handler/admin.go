package handler

import (
	"errors"
	"net/http"

	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/page"
	"github.com/go-sum/server/apperr"
	"github.com/go-sum/server/route"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// AdminElevateForm renders the admin elevation page. Returns 404 when an admin
// already exists so the route is effectively hidden.
func (h *Handler) AdminElevateForm(c *echo.Context) error {
	ctx := c.Request().Context()
	req := h.request(c)

	hasAdmin, err := h.services.User.HasAdmin(ctx)
	if err != nil {
		return apperr.Unavailable("Unable to check admin status right now.", err)
	}
	if hasAdmin {
		return apperr.NotFound("The requested page does not exist.")
	}

	return view.Render(c, req, page.AdminElevatePage(req), nil)
}

// AdminElevate promotes the current user to admin. Only succeeds when no admin
// exists yet; otherwise returns 404.
func (h *Handler) AdminElevate(c *echo.Context) error {
	ctx := c.Request().Context()

	userIDStr, _ := c.Get(h.cfg.App.Keys.UserID).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperr.Unauthorized("Your session is invalid. Please sign in again.")
	}

	_, err = h.services.User.ElevateToAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrAdminExists) {
			return apperr.NotFound("The requested page does not exist.")
		}
		if errors.Is(err, model.ErrUserNotFound) {
			return apperr.Unauthorized("Your account could not be found. Please sign in again.")
		}
		return apperr.Unavailable("Unable to complete elevation right now.", err)
	}

	return c.Redirect(http.StatusSeeOther, route.Reverse(c.Echo().Router().Routes(), "home.show"))
}
