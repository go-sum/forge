package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type adminService interface {
	CountUsers(context.Context) (int64, error)
	ListUsers(context.Context, int, int) ([]model.User, error)
	GetUserByID(context.Context, uuid.UUID) (model.User, error)
	UpdateUser(context.Context, uuid.UUID, model.UpdateUserInput) (model.User, error)
	DeleteUser(context.Context, uuid.UUID) error
	HasAdmin(context.Context) (bool, error)
	ElevateToAdmin(context.Context, uuid.UUID) (model.User, error)
}

const (
	userListDefaultPerPage = 20
	userListMaxPerPage     = 100
)

// AdminHandler owns auth-domain admin/account HTTP workflows.
type AdminHandler struct {
	service   adminService
	forms     FormParser
	redirect  Redirector
	pages     AdminPageRenderer
	homePath  func() string
	requestFn func(c *echo.Context) Request
}

func NewAdminHandler(svc AdminService, cfg AdminHandlerConfig) *AdminHandler {
	return &AdminHandler{
		service:   svc,
		forms:     cfg.Forms,
		redirect:  cfg.Redirect,
		pages:     cfg.Pages,
		homePath:  resolvePath(cfg.HomePath, cfg.HomePathFn),
		requestFn: cfg.RequestFn,
	}
}

func (h *AdminHandler) req(c *echo.Context) Request {
	if h.requestFn != nil {
		return h.requestFn(c)
	}
	return Request{}
}

func (h *AdminHandler) AdminElevatePage(c *echo.Context) error {
	req := h.req(c)
	hasAdmin, err := h.service.HasAdmin(c.Request().Context())
	if err != nil {
		return errUnavailable("Unable to check admin status right now.", err)
	}
	if hasAdmin {
		return errNotFound("The requested page does not exist.")
	}
	return renderOK(c, h.pages.AdminElevatePage(req))
}

func (h *AdminHandler) AdminElevate(c *echo.Context) error {
	userID, err := uuid.Parse(UserID(c))
	if err != nil {
		return errUnauthorized("Your session is invalid. Please sign in again.")
	}

	_, err = h.service.ElevateToAdmin(c.Request().Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrAdminExists):
			return errNotFound("The requested page does not exist.")
		case errors.Is(err, model.ErrUserNotFound):
			return errUnauthorized("Your account could not be found. Please sign in again.")
		default:
			return errUnavailable("Unable to complete elevation right now.", err)
		}
	}

	return h.redirect.Redirect(c.Response(), c.Request(), h.homePath())
}

func (h *AdminHandler) UserList(c *echo.Context) error {
	req := h.req(c)
	pg := newAdminPage(c.Request())

	total, err := h.service.CountUsers(c.Request().Context())
	if err != nil {
		return errUnavailable("Unable to load users right now.", err)
	}
	pg.setTotal(int(total))

	users, err := h.service.ListUsers(c.Request().Context(), pg.Page, pg.PerPage)
	if err != nil {
		return errUnavailable("Unable to load users right now.", err)
	}

	data := AdminUsersPageData{
		Users:      users,
		Page:       pg.Page,
		PerPage:    pg.PerPage,
		TotalItems: pg.TotalItems,
		TotalPages: pg.TotalPages,
	}
	if req.IsPartial() {
		return renderOK(c, h.pages.UserListRegion(req, data))
	}
	return renderOK(c, h.pages.UserListPage(req, data))
}

func (h *AdminHandler) UserEditForm(c *echo.Context) error {
	req := h.req(c)
	user, err := h.loadUser(c)
	if err != nil {
		return err
	}
	return renderOK(c, h.pages.UserEditForm(req, AdminUserFormData{User: user}))
}

func (h *AdminHandler) UserRow(c *echo.Context) error {
	req := h.req(c)
	user, err := h.loadUser(c)
	if err != nil {
		return err
	}
	return renderOK(c, h.pages.UserRow(req, user))
}

func (h *AdminHandler) UserUpdate(c *echo.Context) error {
	req := h.req(c)
	id, err := parseUserID(c)
	if err != nil {
		return err
	}

	var input model.UpdateUserInput
	sub := h.forms.Parse(c, &input)
	if !sub.IsValid() {
		return renderNode(c, http.StatusUnprocessableEntity, h.pages.UserEditForm(req, AdminUserFormData{
			User:   model.User{ID: id},
			Values: input,
			Errors: collectErrors(sub),
		}))
	}

	user, err := h.service.UpdateUser(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, model.ErrEmailTaken) {
			sub.SetFieldError("Email", "Email already in use.")
			return renderNode(c, http.StatusConflict, h.pages.UserEditForm(req, AdminUserFormData{
				User:   model.User{ID: id},
				Values: input,
				Errors: collectErrors(sub),
			}))
		}
		if errors.Is(err, model.ErrUserNotFound) {
			return errNotFound("The requested user does not exist.")
		}
		return errUnavailable("Unable to update this user right now.", err)
	}

	return renderOK(c, h.pages.UserRow(req, user))
}

func (h *AdminHandler) UserDelete(c *echo.Context) error {
	id, err := parseUserID(c)
	if err != nil {
		return err
	}
	if err := h.service.DeleteUser(c.Request().Context(), id); err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return errNotFound("The requested user does not exist.")
		}
		return errUnavailable("Unable to delete this user right now.", err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AdminHandler) loadUser(c *echo.Context) (model.User, error) {
	id, err := parseUserID(c)
	if err != nil {
		return model.User{}, err
	}
	user, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return model.User{}, errNotFound("The requested user does not exist.")
		}
		return model.User{}, errUnavailable("Unable to load this user right now.", err)
	}
	return user, nil
}

func parseUserID(c *echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, errBadRequest("The user ID in the URL is invalid.")
	}
	return id, nil
}

type adminPage struct {
	Page       int
	PerPage    int
	TotalItems int
	TotalPages int
}

func newAdminPage(r *http.Request) adminPage {
	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	perPage := userListDefaultPerPage
	if pp, err := strconv.Atoi(r.URL.Query().Get("per_page")); err == nil && pp > 0 {
		perPage = pp
	}
	if perPage > userListMaxPerPage {
		perPage = userListMaxPerPage
	}
	return adminPage{Page: page, PerPage: perPage}
}

func (p *adminPage) setTotal(total int) {
	p.TotalItems = total
	if p.PerPage <= 0 {
		p.TotalPages = 0
		return
	}
	p.TotalPages = (total + p.PerPage - 1) / p.PerPage
}

func collectErrors(sub FormSubmission) map[string][]string {
	return sub.GetErrors()
}
