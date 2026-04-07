package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	authmodel "github.com/go-sum/auth/model"
	"github.com/go-sum/forge/internal/model"

	"github.com/google/uuid"
)

func TestAdminElevateForm(t *testing.T) {
	tests := []struct {
		name       string
		hasAdminFn func(context.Context) (bool, error)
		wantStatus int
		wantBody   string
	}{
		{
			name: "no admin exists renders form",
			hasAdminFn: func(context.Context) (bool, error) {
				return false, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   "Become Admin",
		},
		{
			name: "admin exists returns 404",
			hasAdminFn: func(context.Context) (bool, error) {
				return true, nil
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "HasAdmin error returns 503",
			hasAdminFn: func(context.Context) (bool, error) {
				return false, errors.New("db down")
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(fakeUserService{
				hasAdminFn: tt.hasAdminFn,
			}, nil)
			c, rec := newRequestContext(http.MethodGet, "/account/admin", nil)
			setCSRFToken(c)
			setUserID(c, testUser.ID.String())

			err := h.AdminElevateForm(c)

			if tt.wantStatus == http.StatusOK {
				if err != nil {
					t.Fatalf("AdminElevateForm() error = %v", err)
				}
				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
				}
				body := rec.Body.String()
				if !strings.Contains(body, tt.wantBody) {
					t.Fatalf("body does not contain %q, got %q", tt.wantBody, body)
				}
			} else {
				assertAppErrorStatus(t, err, tt.wantStatus)
			}
		})
	}
}

func TestAdminElevateFormRendersFullPage(t *testing.T) {
	h := newTestHandler(fakeUserService{
		hasAdminFn: func(context.Context) (bool, error) {
			return false, nil
		},
	}, nil)
	c, rec := newRequestContext(http.MethodGet, "/account/admin", nil)
	setCSRFToken(c)
	setUserID(c, testUser.ID.String())

	if err := h.AdminElevateForm(c); err != nil {
		t.Fatalf("AdminElevateForm() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<html") {
		t.Fatalf("expected full page with <html, got %q", body)
	}
	if !strings.Contains(body, "Elevate to Admin") {
		t.Fatalf("body does not contain submit button text, got %q", body)
	}
}

func TestAdminElevate(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		elevateToAdminFn func(context.Context, uuid.UUID) (authmodel.User, error)
		wantStatus       int
		wantLocation     string
	}{
		{
			name:   "happy path redirects to home",
			userID: testUser.ID.String(),
			elevateToAdminFn: func(_ context.Context, id uuid.UUID) (authmodel.User, error) {
				if id != testUser.ID {
					t.Fatalf("ElevateToAdmin id = %s, want %s", id, testUser.ID)
				}
				elevated := testUser
				elevated.Role = authmodel.RoleAdmin
				return elevated, nil
			},
			wantStatus:   http.StatusSeeOther,
			wantLocation: "/",
		},
		{
			name:   "admin exists returns 404",
			userID: testUser.ID.String(),
			elevateToAdminFn: func(context.Context, uuid.UUID) (authmodel.User, error) {
				return authmodel.User{}, model.ErrAdminExists
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "user not found returns 401",
			userID: testUser.ID.String(),
			elevateToAdminFn: func(context.Context, uuid.UUID) (authmodel.User, error) {
				return authmodel.User{}, authmodel.ErrUserNotFound
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "service error returns 503",
			userID: testUser.ID.String(),
			elevateToAdminFn: func(context.Context, uuid.UUID) (authmodel.User, error) {
				return authmodel.User{}, errors.New("unexpected failure")
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "missing user ID returns 401",
			userID:     "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid user ID returns 401",
			userID:     "not-a-uuid",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(fakeUserService{
				elevateToAdminFn: tt.elevateToAdminFn,
			}, nil)
			c, rec := newRequestContext(http.MethodPost, "/account/admin", nil)
			if tt.userID != "" {
				setUserID(c, tt.userID)
			}

			err := h.AdminElevate(c)

			if tt.wantStatus == http.StatusSeeOther {
				if err != nil {
					t.Fatalf("AdminElevate() error = %v", err)
				}
				if rec.Code != http.StatusSeeOther {
					t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
				}
				location := rec.Header().Get("Location")
				if location != tt.wantLocation {
					t.Fatalf("Location = %q, want %q", location, tt.wantLocation)
				}
			} else {
				assertAppErrorStatus(t, err, tt.wantStatus)
			}
		})
	}
}
