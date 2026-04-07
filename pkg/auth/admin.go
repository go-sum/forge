package auth

import (
	"context"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	g "maragu.dev/gomponents"
)

// AdminService owns auth-domain administrative user workflows.
type AdminService interface {
	CountUsers(context.Context) (int64, error)
	ListUsers(context.Context, int, int) ([]model.User, error)
	GetUserByID(context.Context, uuid.UUID) (model.User, error)
	UpdateUser(context.Context, uuid.UUID, model.UpdateUserInput) (model.User, error)
	DeleteUser(context.Context, uuid.UUID) error
	HasAdmin(context.Context) (bool, error)
	ElevateToAdmin(context.Context, uuid.UUID) (model.User, error)
}

// AdminUsersPageData drives the app-owned user-management views rendered by the host.
type AdminUsersPageData struct {
	Users      []model.User
	Page       int
	PerPage    int
	TotalItems int
	TotalPages int
}

// AdminUserFormData drives the app-owned inline user edit form view.
type AdminUserFormData struct {
	User   model.User
	Values model.UpdateUserInput
	Errors map[string][]string
}

// AdminPageRenderer renders app-owned admin/account views for auth-owned handlers.
type AdminPageRenderer interface {
	AdminElevatePage(req Request) g.Node
	UserListPage(req Request, data AdminUsersPageData) g.Node
	UserListRegion(req Request, data AdminUsersPageData) g.Node
	UserEditForm(req Request, data AdminUserFormData) g.Node
	UserRow(req Request, user model.User) g.Node
}

// AdminHandlerConfig parameterises the auth admin handler with host-owned dependencies.
type AdminHandlerConfig struct {
	Forms      FormParser
	Redirect   Redirector
	Pages      AdminPageRenderer
	HomePath   string
	HomePathFn func() string
	RequestFn  func(c *echo.Context) Request
}
