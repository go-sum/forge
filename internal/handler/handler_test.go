package handler

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"starter/internal/apperr"
	"starter/internal/model"
	"starter/pkg/auth"
	"starter/pkg/ctxkeys"
	"starter/pkg/validate"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

const testCSRFToken = "csrf-token"

var testUser = model.User{
	ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	Email:       "ada@example.com",
	DisplayName: "Ada Lovelace",
	Role:        "admin",
	CreatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
	UpdatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
}

type fakeAuthService struct {
	loginFn    func(context.Context, model.LoginInput) (model.User, error)
	registerFn func(context.Context, model.CreateUserInput) (model.User, error)
}

func (f fakeAuthService) Login(ctx context.Context, input model.LoginInput) (model.User, error) {
	if f.loginFn != nil {
		return f.loginFn(ctx, input)
	}
	return model.User{}, errors.New("unexpected Login call")
}

func (f fakeAuthService) Register(ctx context.Context, input model.CreateUserInput) (model.User, error) {
	if f.registerFn != nil {
		return f.registerFn(ctx, input)
	}
	return model.User{}, errors.New("unexpected Register call")
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

func newTestHandler(authSvc authService, userSvc userService, checkHealth func(context.Context) error) *Handler {
	if checkHealth == nil {
		checkHealth = func(context.Context) error { return nil }
	}
	return &Handler{
		services: handlerServices{
			Auth: authSvc,
			User: userSvc,
		},
		sessions:    auth.NewSessionStore(testSessionConfig()),
		validator:   validate.New(),
		checkHealth: checkHealth,
	}
}

func testSessionConfig() auth.SessionConfig {
	return auth.SessionConfig{
		Name:       "test-session",
		AuthKey:    strings.Repeat("a", 32),
		EncryptKey: strings.Repeat("b", 32),
		MaxAge:     3600,
	}
}

func newRequestContext(method, target string, body io.Reader) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
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
	c.Set(string(ctxkeys.UserID), userID)
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
