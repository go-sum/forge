package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authadapter "github.com/go-sum/auth/adapters/echocomponentry"
	"github.com/go-sum/auth/session"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/handler"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/repository"
	"github.com/go-sum/forge/internal/service"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/server/validate"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

func TestRegisterRoutesLoadsUserContextForPublicPages(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{
		App: config.AppConfig{
			Security: config.SecurityConfig{
				ExternalOrigin: "http://localhost:3000",
				CSRF: config.CSRFConfig{
					FormField:  "_csrf",
					HeaderName: "X-CSRF-Token",
				},
			},
			Auth: config.AuthConfig{
				Session: config.SessionConfig{
					Name:       "_session",
					AuthKey:    "12345678901234567890123456789012",
					EncryptKey: "12345678901234567890123456789012",
				},
			},
			Keys: config.ContextKeysConfig{
				UserID:      "user_id",
				UserRole:    "user_role",
				DisplayName: "user_display_name",
				CSRF:        "csrf",
			},
		},
		Nav: config.NavConfig{
			Brand: config.NavbarBrand{Label: "Starter", Href: "/"},
			Sections: []config.NavSection{
				{Items: []config.NavItem{{Label: "Home", Href: "/"}}},
				{Align: "end", Items: []config.NavItem{
					{
						Label: "Account",
						Items: []config.NavItem{
							{Slot: "user_name", Visibility: "user"},
							{Slot: "signout", Label: "Signout", Visibility: "user"},
							{Label: "Sign In", Href: "/signin", Visibility: "guest"},
							{Label: "Sign Up", Href: "/signup", Visibility: "guest"},
						},
					},
				}},
			},
		},
	}

	sessions, err := session.NewSessionStore(session.SessionConfig{
		Name:       cfg.App.Auth.Session.Name,
		AuthKey:    cfg.App.Auth.Session.AuthKey,
		EncryptKey: cfg.App.Auth.Session.EncryptKey,
	})
	if err != nil {
		t.Fatalf("NewSessionStore() error = %v", err)
	}

	container := &Container{
		Config:    cfg,
		Web:       e,
		Sessions:  sessions,
		Validator: validate.New(),
		Repos: &repository.Repositories{
			User: routesTestUserRepo{
				users: map[uuid.UUID]model.User{
					uuid.MustParse("11111111-1111-1111-1111-111111111111"): {
						ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
						Email:       "ada@example.com",
						DisplayName: "Ada Lovelace",
						Role:        "admin",
					},
				},
			},
		},
	}

	h := handler.New(&service.Services{}, container.Validator, func(context.Context) error { return nil }, cfg, func() echo.Routes {
		return e.Router().Routes()
	})
	authH := authadapter.New(nil, sessions, container.Validator, authadapter.Config{
		CSRFField:       cfg.App.Security.CSRF.FormField,
		SigninPath:      "/signin",
		SignupPath:      "/signup",
		VerifyPath:      "/verify",
		VerifyURL:       "http://localhost:3000/verify",
		EmailChangePath: "/account/email",
		HomePath:        "/",
		RequestFn: func(ec *echo.Context) authadapter.Request {
			req := view.NewRequest(ec, cfg)
			return authadapter.Request{
				CSRFToken: req.CSRFToken,
				PageFn:    req.Page,
			}
		},
	})

	if err := RegisterRoutes(container, h, authH); err != nil {
		t.Fatalf("RegisterRoutes() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	cookieRec := httptest.NewRecorder()
	if err := sessions.SetUserID(cookieRec, req, "11111111-1111-1111-1111-111111111111"); err != nil {
		t.Fatalf("SetUserID() error = %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range cookieRec.Result().Cookies() {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Ada Lovelace") {
		t.Fatalf("body missing display name: %s", rec.Body.String())
	}
}

type routesTestUserRepo struct {
	users map[uuid.UUID]model.User
}

func (r routesTestUserRepo) Create(context.Context, string, string, string, bool) (model.User, error) {
	return model.User{}, nil
}

func (r routesTestUserRepo) GetByID(_ context.Context, id uuid.UUID) (model.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return model.User{}, model.ErrUserNotFound
}

func (r routesTestUserRepo) GetByEmail(context.Context, string) (model.User, error) {
	return model.User{}, nil
}

func (r routesTestUserRepo) List(context.Context, int32, int32) ([]model.User, error) {
	return nil, nil
}

func (r routesTestUserRepo) Update(context.Context, uuid.UUID, string, string, string) (model.User, error) {
	return model.User{}, nil
}

func (r routesTestUserRepo) UpdateEmail(context.Context, uuid.UUID, string) (model.User, error) {
	return model.User{}, nil
}

func (r routesTestUserRepo) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (r routesTestUserRepo) Count(context.Context) (int64, error) {
	return int64(len(r.users)), nil
}

var _ repository.UserRepository = routesTestUserRepo{}

func TestRegisterRoutes_AccessTiers(t *testing.T) {
	const (
		adminID   = "11111111-1111-1111-1111-111111111111"
		regularID = "22222222-2222-2222-2222-222222222222"
	)

	e := echo.New()
	cfg := &config.Config{
		App: config.AppConfig{
			Security: config.SecurityConfig{
				ExternalOrigin: "http://localhost:3000",
				CSRF: config.CSRFConfig{
					FormField:  "_csrf",
					HeaderName: "X-CSRF-Token",
				},
			},
			Auth: config.AuthConfig{
				Session: config.SessionConfig{
					Name:       "_session",
					AuthKey:    "12345678901234567890123456789012",
					EncryptKey: "12345678901234567890123456789012",
				},
			},
			Keys: config.ContextKeysConfig{
				UserID:      "user_id",
				UserRole:    "user_role",
				DisplayName: "user_display_name",
				CSRF:        "csrf",
			},
		},
		Nav: config.NavConfig{
			Brand: config.NavbarBrand{Label: "Starter", Href: "/"},
		},
	}

	sessions, err := session.NewSessionStore(session.SessionConfig{
		Name:       cfg.App.Auth.Session.Name,
		AuthKey:    cfg.App.Auth.Session.AuthKey,
		EncryptKey: cfg.App.Auth.Session.EncryptKey,
	})
	if err != nil {
		t.Fatalf("NewSessionStore() error = %v", err)
	}

	repo := routesTestUserRepo{
		users: map[uuid.UUID]model.User{
			uuid.MustParse(adminID): {
				ID:          uuid.MustParse(adminID),
				Email:       "admin@example.com",
				DisplayName: "Admin User",
				Role:        "admin",
			},
			uuid.MustParse(regularID): {
				ID:          uuid.MustParse(regularID),
				Email:       "user@example.com",
				DisplayName: "Regular User",
				Role:        "user",
			},
		},
	}

	container := &Container{
		Config:    cfg,
		Web:       e,
		Sessions:  sessions,
		Validator: validate.New(),
		Repos:     &repository.Repositories{User: repo},
	}

	h := handler.New(
		&service.Services{User: service.NewUserService(repo)},
		container.Validator,
		func(context.Context) error { return nil },
		cfg,
		func() echo.Routes { return e.Router().Routes() },
	)
	authH := authadapter.New(nil, sessions, container.Validator, authadapter.Config{
		CSRFField:       cfg.App.Security.CSRF.FormField,
		SigninPath:      "/signin",
		SignupPath:      "/signup",
		VerifyPath:      "/verify",
		VerifyURL:       "http://localhost:3000/verify",
		EmailChangePath: "/account/email",
		HomePath:        "/",
		RequestFn: func(ec *echo.Context) authadapter.Request {
			req := view.NewRequest(ec, cfg)
			return authadapter.Request{
				CSRFToken: req.CSRFToken,
				PageFn:    req.Page,
			}
		},
	})

	if err := RegisterRoutes(container, h, authH); err != nil {
		t.Fatalf("RegisterRoutes() error = %v", err)
	}

	// makeCookie returns a session cookie for the given user ID.
	makeCookie := func(t *testing.T, userID string) *http.Cookie {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if err := sessions.SetUserID(rec, req, userID); err != nil {
			t.Fatalf("SetUserID() error = %v", err)
		}
		cookies := rec.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("SetUserID() produced no cookies")
		}
		return cookies[0]
	}

	tests := []struct {
		name         string
		path         string
		cookie       *http.Cookie
		wantStatus   int
		wantLocation string
	}{
		{
			name:       "public route allows unauthenticated",
			path:       "/",
			wantStatus: http.StatusOK,
		},
		{
			name:         "admin route redirects unauthenticated to signin",
			path:         "/users",
			wantStatus:   http.StatusSeeOther,
			wantLocation: "/signin",
		},
		{
			name:       "admin route rejects authenticated non-admin with 403",
			path:       "/users",
			cookie:     makeCookie(t, regularID),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "admin route allows authenticated admin",
			path:       "/users",
			cookie:     makeCookie(t, adminID),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d\nbody: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantLocation != "" {
				if got := rec.Header().Get("Location"); got != tt.wantLocation {
					t.Errorf("Location = %q, want %q", got, tt.wantLocation)
				}
			}
		})
	}
}
