package auth

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/go-sum/auth/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type fakeAdminService struct {
	countFn   func(context.Context) (int64, error)
	listFn    func(context.Context, int, int) ([]model.User, error)
	getByIDFn func(context.Context, uuid.UUID) (model.User, error)
	updateFn  func(context.Context, uuid.UUID, model.UpdateUserInput) (model.User, error)
	deleteFn  func(context.Context, uuid.UUID) error
	hasAdmin  func(context.Context) (bool, error)
	elevateFn func(context.Context, uuid.UUID) (model.User, error)
}

func (f fakeAdminService) CountUsers(ctx context.Context) (int64, error) {
	if f.countFn != nil {
		return f.countFn(ctx)
	}
	return 0, errors.New("unexpected CountUsers call")
}

func (f fakeAdminService) ListUsers(ctx context.Context, page, perPage int) ([]model.User, error) {
	if f.listFn != nil {
		return f.listFn(ctx, page, perPage)
	}
	return nil, errors.New("unexpected ListUsers call")
}

func (f fakeAdminService) GetUserByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return model.User{}, errors.New("unexpected GetUserByID call")
}

func (f fakeAdminService) UpdateUser(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (model.User, error) {
	if f.updateFn != nil {
		return f.updateFn(ctx, id, input)
	}
	return model.User{}, errors.New("unexpected UpdateUser call")
}

func (f fakeAdminService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return errors.New("unexpected DeleteUser call")
}

func (f fakeAdminService) HasAdmin(ctx context.Context) (bool, error) {
	if f.hasAdmin != nil {
		return f.hasAdmin(ctx)
	}
	return false, errors.New("unexpected HasAdmin call")
}

func (f fakeAdminService) ElevateToAdmin(ctx context.Context, id uuid.UUID) (model.User, error) {
	if f.elevateFn != nil {
		return f.elevateFn(ctx, id)
	}
	return model.User{}, errors.New("unexpected ElevateToAdmin call")
}

type fakeAdminPageRenderer struct{}

func (r *fakeAdminPageRenderer) AdminElevatePage(req Request) g.Node {
	return h.Div(
		h.Input(h.Type("hidden"), h.Name(req.CSRFFieldName), h.Value(req.CSRFToken)),
		g.Text("Become Admin"),
	)
}

func (r *fakeAdminPageRenderer) UserListPage(_ Request, data AdminUsersPageData) g.Node {
	nodes := []g.Node{g.Text("Users")}
	for _, user := range data.Users {
		nodes = append(nodes, g.Text(user.Email))
	}
	return h.Div(nodes...)
}

func (r *fakeAdminPageRenderer) UserListRegion(_ Request, data AdminUsersPageData) g.Node {
	nodes := []g.Node{h.Div(h.ID("users-list-region"))}
	for _, user := range data.Users {
		nodes = append(nodes, g.Text(user.Email))
	}
	return h.Div(nodes...)
}

func (r *fakeAdminPageRenderer) UserEditForm(_ Request, data AdminUserFormData) g.Node {
	nodes := []g.Node{h.Input(h.Name("email"), h.Value(data.Values.Email))}
	for _, msg := range data.Errors["Email"] {
		nodes = append(nodes, g.Text(msg))
	}
	return h.Div(nodes...)
}

func (r *fakeAdminPageRenderer) UserRow(_ Request, user model.User) g.Node {
	return h.Div(g.Text(user.DisplayName))
}

func newAdminHandler(svc AdminService) *AdminHandler {
	return NewAdminHandler(svc, AdminHandlerConfig{
		Forms:    &fakeFormParser{},
		Redirect: &fakeRedirector{},
		Pages:    &fakeAdminPageRenderer{},
		HomePath: "/",
		RequestFn: func(_ *echo.Context) Request {
			return Request{CSRFToken: testCSRFToken, CSRFFieldName: "_csrf"}
		},
	})
}

func TestAdminElevatePageRenders(t *testing.T) {
	h := newAdminHandler(fakeAdminService{
		hasAdmin: func(context.Context) (bool, error) { return false, nil },
	})
	c, rec := newRequestContext(http.MethodGet, "/account/admin", nil)

	if err := h.AdminElevatePage(c); err != nil {
		t.Fatalf("AdminElevatePage() error = %v", err)
	}
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Become Admin") {
		t.Fatalf("response = %d %q", rec.Code, rec.Body.String())
	}
}

func TestUserListPartialRendersRegion(t *testing.T) {
	h := NewAdminHandler(fakeAdminService{
		countFn: func(context.Context) (int64, error) { return 1, nil },
		listFn: func(context.Context, int, int) ([]model.User, error) {
			return []model.User{handlerTestUser}, nil
		},
	}, AdminHandlerConfig{
		Forms:    &fakeFormParser{},
		Redirect: &fakeRedirector{},
		Pages:    &fakeAdminPageRenderer{},
		RequestFn: func(_ *echo.Context) Request {
			return Request{Partial: true}
		},
	})
	c, rec := newRequestContext(http.MethodGet, "/users", nil)

	if err := h.UserList(c); err != nil {
		t.Fatalf("UserList() error = %v", err)
	}
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "users-list-region") {
		t.Fatalf("response = %d %q", rec.Code, rec.Body.String())
	}
}

func TestUserUpdateConflictRenders409(t *testing.T) {
	h := newAdminHandler(fakeAdminService{
		updateFn: func(context.Context, uuid.UUID, model.UpdateUserInput) (model.User, error) {
			return model.User{}, model.ErrEmailTaken
		},
	})
	c, rec := newFormContext(http.MethodPut, "/users/"+handlerTestUser.ID.String(), url.Values{
		"email": {"grace@example.com"},
	})
	c.SetPath("/users/:id")
	c.SetPathValues(echo.PathValues{{
		Name:  "id",
		Value: handlerTestUser.ID.String(),
	}})

	if err := h.UserUpdate(c); err != nil {
		t.Fatalf("UserUpdate() error = %v", err)
	}
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "Email already in use.") {
		t.Fatalf("response = %d %q", rec.Code, rec.Body.String())
	}
}

func TestAdminElevateRedirectsHome(t *testing.T) {
	h := newAdminHandler(fakeAdminService{
		elevateFn: func(_ context.Context, id uuid.UUID) (model.User, error) {
			if id != handlerTestUser.ID {
				t.Fatalf("id = %s", id)
			}
			return handlerTestUser, nil
		},
	})
	c, rec := newRequestContext(http.MethodPost, "/account/admin", nil)
	c.Set(ContextKeyUserID, handlerTestUser.ID.String())

	if err := h.AdminElevate(c); err != nil {
		t.Fatalf("AdminElevate() error = %v", err)
	}
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Fatalf("response = %d location=%q", rec.Code, rec.Header().Get("Location"))
	}
}
