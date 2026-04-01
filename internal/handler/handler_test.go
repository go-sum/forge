package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/server/apperr"
	"github.com/go-sum/server/route"
	"github.com/go-sum/server/validate"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

const testCSRFToken = "csrf-token"

var testKeys = config.ContextKeysConfig{
	UserID:      "user_id",
	UserRole:    "user_role",
	DisplayName: "user_display_name",
	CSRF:        "csrf",
}

var testUser = model.User{
	ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	Email:       "ada@example.com",
	DisplayName: "Ada Lovelace",
	Role:        "admin",
	CreatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
	UpdatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
}

type fakeUserService struct {
	countFn  func(context.Context) (int64, error)
	listFn   func(context.Context, int, int) ([]model.User, error)
	getByID  func(context.Context, uuid.UUID) (model.User, error)
	updateFn func(context.Context, uuid.UUID, model.UpdateUserInput) (model.User, error)
	deleteFn func(context.Context, uuid.UUID) error
}

func (f fakeUserService) Count(ctx context.Context) (int64, error) {
	if f.countFn != nil {
		return f.countFn(ctx)
	}
	return 0, errors.New("unexpected Count call")
}

func (f fakeUserService) List(ctx context.Context, page, perPage int) ([]model.User, error) {
	if f.listFn != nil {
		return f.listFn(ctx, page, perPage)
	}
	return nil, errors.New("unexpected List call")
}

func (f fakeUserService) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	if f.getByID != nil {
		return f.getByID(ctx, id)
	}
	return model.User{}, errors.New("unexpected GetByID call")
}

func (f fakeUserService) Update(ctx context.Context, id uuid.UUID, input model.UpdateUserInput) (model.User, error) {
	if f.updateFn != nil {
		return f.updateFn(ctx, id, input)
	}
	return model.User{}, errors.New("unexpected Update call")
}

func (f fakeUserService) Delete(ctx context.Context, id uuid.UUID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return errors.New("unexpected Delete call")
}

func newTestHandler(userSvc userService, _ ...func(context.Context) error) *Handler {
	e := echo.New()
	registerTestRoutes(e)
	return &Handler{
		services: handlerServices{
			User: userSvc,
		},
		validator: validate.New(),
		cfg:       &config.Config{App: config.AppConfig{Keys: testKeys}},
		routes:    func() echo.Routes { return e.Router().Routes() },
	}
}

func newRequestContext(method, target string, body io.Reader) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	registerTestRoutes(e)
	req := httptest.NewRequest(method, target, body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func newFormContext(method, target string, values url.Values) (*echo.Context, *httptest.ResponseRecorder) {
	c, rec := newRequestContext(method, target, strings.NewReader(values.Encode()))
	c.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	return c, rec
}

func setCSRFToken(c *echo.Context) {
	c.Set(middleware.DefaultCSRFConfig.ContextKey, testCSRFToken)
}

func setUserID(c *echo.Context, userID string) {
	c.Set(testKeys.UserID, userID)
}

func setPathParam(c *echo.Context, path, name, value string) {
	c.SetPath(path)
	c.SetPathValues(echo.PathValues{{
		Name:  name,
		Value: value,
	}})
}

func assertAppErrorStatus(t *testing.T, err error, status int) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %T, want *apperr.Error", err)
	}
	if appErr.Status != status {
		t.Fatalf("status = %d, want %d", appErr.Status, status)
	}
}

func registerTestRoutes(e *echo.Echo) {
	noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "components.list", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signup", Name: "signup.get", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/robots.txt", Name: "robots.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/sitemap.xml", Name: "sitemap.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/contact", Name: "contact.show", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodPost, Path: "/contact", Name: "contact.submit", Handler: noOp})

	users := e.Group("/users")
	route.Add(users, echo.Route{Method: http.MethodGet, Path: "", Name: "user.list", Handler: noOp})
	route.Add(users, echo.Route{Method: http.MethodGet, Path: "/:id/edit", Name: "user.edit", Handler: noOp})
	route.Add(users, echo.Route{Method: http.MethodGet, Path: "/:id/row", Name: "user.row", Handler: noOp})
	route.Add(users, echo.Route{Method: http.MethodPut, Path: "/:id", Name: "user.update", Handler: noOp})
	route.Add(users, echo.Route{Method: http.MethodDelete, Path: "/:id", Name: "user.delete", Handler: noOp})
}
