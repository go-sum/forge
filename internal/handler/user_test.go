package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"starter/internal/model"
	"starter/internal/routes"

	"github.com/google/uuid"
)

func TestUserListRendersFullPage(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{
		countFn: func(context.Context) (int64, error) { return 21, nil },
		listFn: func(_ context.Context, page, perPage int) ([]model.User, error) {
			if page != 2 || perPage != 10 {
				t.Fatalf("page=%d perPage=%d", page, perPage)
			}
			return []model.User{testUser}, nil
		},
	}, nil)
	c, rec := newRequestContext(http.MethodGet, routes.Users+"?page=2&per_page=10", nil)
	setCSRFToken(c)
	setUserID(c, testUser.ID.String())

	if err := h.UserList(c); err != nil {
		t.Fatalf("UserList() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<html") || !strings.Contains(body, "Users") || !strings.Contains(body, testUser.Email) {
		t.Fatalf("body = %q", body)
	}
}

func TestUserListRendersHTMXFragment(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{
		countFn: func(context.Context) (int64, error) { return 1, nil },
		listFn:  func(context.Context, int, int) ([]model.User, error) { return []model.User{testUser}, nil },
	}, nil)
	c, rec := newRequestContext(http.MethodGet, routes.Users, nil)
	c.Request().Header.Set("HX-Request", "true")

	if err := h.UserList(c); err != nil {
		t.Fatalf("UserList() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "<html") || !strings.Contains(body, `id="users-list-region"`) {
		t.Fatalf("body = %q", body)
	}
}

func TestUserListCountFailureReturnsUnavailable(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{
		countFn: func(context.Context) (int64, error) { return 0, errors.New("db down") },
	}, nil)
	c, _ := newRequestContext(http.MethodGet, routes.Users, nil)

	err := h.UserList(c)
	assertAppErrorStatus(t, err, http.StatusServiceUnavailable)
}

func TestUserEditFormRejectsInvalidID(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{}, nil)
	c, _ := newRequestContext(http.MethodGet, "/users/not-a-uuid/edit", nil)
	setPathParam(c, routes.UserEdit, "id", "not-a-uuid")

	err := h.UserEditForm(c)
	assertAppErrorStatus(t, err, http.StatusBadRequest)
}

func TestUserEditFormRenders(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{
		getByID: func(_ context.Context, id uuid.UUID) (model.User, error) {
			if id != testUser.ID {
				t.Fatalf("id = %s", id)
			}
			return testUser, nil
		},
	}, nil)
	c, rec := newRequestContext(http.MethodGet, routes.UserEditPath(testUser.ID.String()), nil)
	setCSRFToken(c)
	setPathParam(c, routes.UserEdit, "id", testUser.ID.String())

	if err := h.UserEditForm(c); err != nil {
		t.Fatalf("UserEditForm() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `name="display_name"`) || !strings.Contains(body, testUser.Email) {
		t.Fatalf("body = %q", body)
	}
}

func TestUserRowReturnsResolvedError(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{
		getByID: func(context.Context, uuid.UUID) (model.User, error) { return model.User{}, model.ErrUserNotFound },
	}, nil)
	c, _ := newRequestContext(http.MethodGet, routes.UserRowPath(testUser.ID.String()), nil)
	setPathParam(c, routes.UserRow, "id", testUser.ID.String())

	err := h.UserRow(c)
	assertAppErrorStatus(t, err, http.StatusNotFound)
}

func TestUserUpdateValidationFailureRenders422(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{
		getByID: func(context.Context, uuid.UUID) (model.User, error) { return testUser, nil },
	}, nil)
	c, rec := newFormContext(http.MethodPut, routes.UserPath(testUser.ID.String()), url.Values{
		"email": {"not-an-email"},
	})
	setCSRFToken(c)
	setPathParam(c, routes.UserByID, "id", testUser.ID.String())

	if err := h.UserUpdate(c); err != nil {
		t.Fatalf("UserUpdate() error = %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `name="email"`) || !strings.Contains(body, `value="not-an-email"`) {
		t.Fatalf("body = %q", body)
	}
}

func TestUserUpdateConflictRenders409(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{
		updateFn: func(context.Context, uuid.UUID, model.UpdateUserInput) (model.User, error) {
			return model.User{}, model.ErrEmailTaken
		},
		getByID: func(context.Context, uuid.UUID) (model.User, error) { return testUser, nil },
	}, nil)
	c, rec := newFormContext(http.MethodPut, routes.UserPath(testUser.ID.String()), url.Values{
		"email":        {"grace@example.com"},
		"display_name": {"Ada"},
		"role":         {"admin"},
	})
	setCSRFToken(c)
	setPathParam(c, routes.UserByID, "id", testUser.ID.String())

	if err := h.UserUpdate(c); err != nil {
		t.Fatalf("UserUpdate() error = %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Email already in use.") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestUserUpdateRendersUpdatedRow(t *testing.T) {
	updated := testUser
	updated.DisplayName = "Grace Hopper"
	h := newTestHandler(fakeAuthService{}, fakeUserService{
		updateFn: func(_ context.Context, id uuid.UUID, input model.UpdateUserInput) (model.User, error) {
			if id != testUser.ID || input.DisplayName != "Grace Hopper" || input.Role != "admin" {
				t.Fatalf("id=%s input=%#v", id, input)
			}
			return updated, nil
		},
	}, nil)
	c, rec := newFormContext(http.MethodPut, routes.UserPath(testUser.ID.String()), url.Values{
		"display_name": {"Grace Hopper"},
		"role":         {"admin"},
	})
	setPathParam(c, routes.UserByID, "id", testUser.ID.String())

	if err := h.UserUpdate(c); err != nil {
		t.Fatalf("UserUpdate() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Grace Hopper") || !strings.Contains(body, "Delete Grace Hopper?") {
		t.Fatalf("body = %q", body)
	}
}

func TestUserDeleteRejectsInvalidID(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{}, nil)
	c, _ := newRequestContext(http.MethodDelete, "/users/not-a-uuid", nil)
	setPathParam(c, routes.UserByID, "id", "not-a-uuid")

	err := h.UserDelete(c)
	assertAppErrorStatus(t, err, http.StatusBadRequest)
}

func TestUserDeleteReturnsNoContent(t *testing.T) {
	h := newTestHandler(fakeAuthService{}, fakeUserService{
		deleteFn: func(_ context.Context, id uuid.UUID) error {
			if id != testUser.ID {
				t.Fatalf("id = %s", id)
			}
			return nil
		},
	}, nil)
	c, rec := newRequestContext(http.MethodDelete, routes.UserPath(testUser.ID.String()), nil)
	setPathParam(c, routes.UserByID, "id", testUser.ID.String())

	if err := h.UserDelete(c); err != nil {
		t.Fatalf("UserDelete() error = %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}
