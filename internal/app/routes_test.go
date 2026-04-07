package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authmodel "github.com/go-sum/auth/model"
	auth "github.com/go-sum/auth"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/adapters/authview"
	"github.com/go-sum/forge/internal/handler"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/go-sum/forge/internal/service"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/server/validate"
	"github.com/go-sum/session"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

func TestRegisterRoutesSkipsUserHydrationForPublicPages(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{
		App: config.AppConfig{
			Security: config.SecurityConfig{
				ExternalOrigin: "http://localhost:3000",
				CSRF: config.CSRFConfig{
					ContextKey: "csrf",
					FormField:  "_csrf",
					HeaderName: "X-CSRF-Token",
				},
			},
			Session: config.SessionConfig{
				Name:       "_session",
				AuthKey:    "12345678901234567890123456789012",
				EncryptKey: "12345678901234567890123456789012",
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

	sessions, err := session.NewManager(session.Config{
		CookieName: cfg.App.Session.Name,
		AuthKey:    cfg.App.Session.AuthKey,
		EncryptKey: cfg.App.Session.EncryptKey,
	})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	store := &routesTestStore{
		users: map[uuid.UUID]authmodel.User{
			uuid.MustParse("11111111-1111-1111-1111-111111111111"): {
				ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				Email:       "ada@example.com",
				DisplayName: "Ada Lovelace",
				Role:        "admin",
			},
		},
	}

	container := &Container{
		Config:       cfg,
		Web:          e,
		RateLimiters: appserver.NewRateLimiters(cfg),
		Sessions:     sessions,
		Validator:    validate.New(),
		AuthStore:    store,
	}

	h := handler.New(cfg, func() echo.Routes {
		return e.Router().Routes()
	}, &service.Services{}, container.Validator)
	availabilityH := handler.NewAvailability(func(context.Context) error { return nil }, nil, "")
	authH := auth.NewHandler(nil, auth.HandlerConfig{
		Sessions:        &sessionManagerAdapter{mgr: sessions},
		Forms:           &formParserAdapter{v: container.Validator},
		Flash:           &flashAdapter{},
		Redirect:        &redirectAdapter{},
		Pages:           authview.NewRenderer(),
		CSRFField:       cfg.App.Security.CSRF.FormField,
		SigninPath:      "/signin",
		SignupPath:      "/signup",
		VerifyPath:      "/verify",
		VerifyURL:       "http://localhost:3000/verify",
		EmailChangePath: "/account/email",
		HomePath:        "/",
		RequestFn: func(ec *echo.Context) auth.Request {
			req := view.NewRequest(ec, cfg)
			return auth.Request{
				CSRFToken: req.CSRFToken,
				PageFn:    req.Page,
			}
		},
	})

	if err := RegisterRoutes(container, h, availabilityH, authH); err != nil {
		t.Fatalf("RegisterRoutes() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	cookieRec := httptest.NewRecorder()
	state, err := sessions.Load(req)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := state.Put("auth.user_id", "11111111-1111-1111-1111-111111111111"); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if err := sessions.Commit(cookieRec, req, state); err != nil {
		t.Fatalf("Commit() error = %v", err)
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
	if store.getByIDCalls != 0 {
		t.Fatalf("GetByID() calls = %d, want 0 for public page", store.getByIDCalls)
	}
	if !strings.Contains(rec.Body.String(), "Account") {
		t.Fatalf("body missing generic auth label: %s", rec.Body.String())
	}
}

// routesTestStore implements AuthStore (authrepo.UserStore + authrepo.AdminStore) for route tests.
type routesTestStore struct {
	users        map[uuid.UUID]authmodel.User
	getByIDCalls int
}

func (s *routesTestStore) GetByID(_ context.Context, id uuid.UUID) (authmodel.User, error) {
	s.getByIDCalls++
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return authmodel.User{}, authmodel.ErrUserNotFound
}

func (s *routesTestStore) GetByEmail(_ context.Context, email string) (authmodel.User, error) {
	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return authmodel.User{}, authmodel.ErrUserNotFound
}

func (s *routesTestStore) Create(_ context.Context, email, displayName, role string, verified bool) (authmodel.User, error) {
	return authmodel.User{}, nil
}

func (s *routesTestStore) UpdateEmail(_ context.Context, id uuid.UUID, email string) (authmodel.User, error) {
	return authmodel.User{}, nil
}

func (s *routesTestStore) List(_ context.Context, _, _ int32) ([]authmodel.User, error) {
	return nil, nil
}

func (s *routesTestStore) Update(_ context.Context, _ uuid.UUID, _, _, _ string) (authmodel.User, error) {
	return authmodel.User{}, nil
}

func (s *routesTestStore) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *routesTestStore) Count(_ context.Context) (int64, error) {
	return int64(len(s.users)), nil
}

func (s *routesTestStore) HasAdmin(_ context.Context) (bool, error) {
	return false, nil
}

var _ AuthStore = (*routesTestStore)(nil)

// newTestApp builds a minimal application container with fake repositories
// suitable for route integration tests. The returned Echo instance has all
// routes registered and is ready for ServeHTTP calls.
func newTestApp(t *testing.T) (*echo.Echo, session.Manager, *routesTestStore) {
	t.Helper()
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
					ContextKey: "csrf",
					FormField:  "_csrf",
					HeaderName: "X-CSRF-Token",
				},
			},
			Session: config.SessionConfig{
				Name:       "_session",
				AuthKey:    "12345678901234567890123456789012",
				EncryptKey: "12345678901234567890123456789012",
			},
		},
		Nav: config.NavConfig{
			Brand: config.NavbarBrand{Label: "Starter", Href: "/"},
		},
	}

	sessions, err := session.NewManager(session.Config{
		CookieName: cfg.App.Session.Name,
		AuthKey:    cfg.App.Session.AuthKey,
		EncryptKey: cfg.App.Session.EncryptKey,
	})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	store := &routesTestStore{
		users: map[uuid.UUID]authmodel.User{
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
		Config:       cfg,
		Web:          e,
		RateLimiters: appserver.NewRateLimiters(cfg),
		Sessions:     sessions,
		Validator:    validate.New(),
		AuthStore:    store,
	}

	h := handler.New(
		cfg,
		func() echo.Routes { return e.Router().Routes() },
		&service.Services{User: service.NewUserService(store)},
		container.Validator,
	)
	availabilityH := handler.NewAvailability(func(context.Context) error { return nil }, nil, "")
	authH := auth.NewHandler(nil, auth.HandlerConfig{
		Sessions:        &sessionManagerAdapter{mgr: sessions},
		Forms:           &formParserAdapter{v: container.Validator},
		Flash:           &flashAdapter{},
		Redirect:        &redirectAdapter{},
		Pages:           authview.NewRenderer(),
		CSRFField:       cfg.App.Security.CSRF.FormField,
		SigninPath:      "/signin",
		SignupPath:      "/signup",
		VerifyPath:      "/verify",
		VerifyURL:       "http://localhost:3000/verify",
		EmailChangePath: "/account/email",
		HomePath:        "/",
		RequestFn: func(ec *echo.Context) auth.Request {
			req := view.NewRequest(ec, cfg)
			return auth.Request{
				CSRFToken: req.CSRFToken,
				PageFn:    req.Page,
			}
		},
	})

	if err := RegisterRoutes(container, h, availabilityH, authH); err != nil {
		t.Fatalf("RegisterRoutes() error = %v", err)
	}

	return e, sessions, store
}

// adminCookie returns a session cookie for the admin user.
func adminCookie(t *testing.T, mgr session.Manager, adminID string) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	state, err := mgr.Load(req)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := state.Put("auth.user_id", adminID); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if err := mgr.Commit(rec, req, state); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("Commit() produced no cookies")
	}
	return cookies[0]
}

// TestUserRowETagCachingCycle verifies the 200 → 304 caching cycle for the
// GET /users/:id/row HTMX fragment endpoint.
//
// First request: expect 200 with an ETag header in the response.
// Second request with matching If-None-Match: expect 304 with empty body.
func TestUserRowETagCachingCycle(t *testing.T) {
	const adminID = "11111111-1111-1111-1111-111111111111"
	e, sessions, _ := newTestApp(t)
	cookie := adminCookie(t, sessions, adminID)

	path := "/users/" + adminID + "/row"

	// First request — expect 200 with an ETag.
	req1 := httptest.NewRequest(http.MethodGet, path, nil)
	req1.AddCookie(cookie)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("first request: status = %d, want 200\nbody: %s", rec1.Code, rec1.Body.String())
	}
	etagVal := rec1.Header().Get("ETag")
	if etagVal == "" {
		t.Fatal("first request: ETag header not set")
	}
	if got := rec1.Header().Get("Cache-Control"); got != "private, must-revalidate" {
		t.Fatalf("first request: Cache-Control = %q, want %q", got, "private, must-revalidate")
	}
	if got := rec1.Header().Get("Vary"); got != "Cookie" {
		t.Fatalf("first request: Vary = %q, want %q", got, "Cookie")
	}

	// Second request with matching If-None-Match — expect 304.
	req2 := httptest.NewRequest(http.MethodGet, path, nil)
	req2.AddCookie(cookie)
	req2.Header.Set("If-None-Match", etagVal)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusNotModified {
		t.Errorf("second request: status = %d, want 304\nbody: %s", rec2.Code, rec2.Body.String())
	}
	if rec2.Body.Len() != 0 {
		t.Errorf("second request: body should be empty for 304, got %q", rec2.Body.String())
	}
	if got := rec2.Header().Get("Cache-Control"); got != "private, must-revalidate" {
		t.Errorf("second request: Cache-Control = %q, want %q", got, "private, must-revalidate")
	}
	if got := rec2.Header().Get("Vary"); got != "Cookie" {
		t.Errorf("second request: Vary = %q, want %q", got, "Cookie")
	}
}

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
					ContextKey: "csrf",
					FormField:  "_csrf",
					HeaderName: "X-CSRF-Token",
				},
			},
			Session: config.SessionConfig{
				Name:       "_session",
				AuthKey:    "12345678901234567890123456789012",
				EncryptKey: "12345678901234567890123456789012",
			},
		},
		Nav: config.NavConfig{
			Brand: config.NavbarBrand{Label: "Starter", Href: "/"},
		},
	}

	sessions, err := session.NewManager(session.Config{
		CookieName: cfg.App.Session.Name,
		AuthKey:    cfg.App.Session.AuthKey,
		EncryptKey: cfg.App.Session.EncryptKey,
	})
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	store := &routesTestStore{
		users: map[uuid.UUID]authmodel.User{
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
		Config:       cfg,
		Web:          e,
		RateLimiters: appserver.NewRateLimiters(cfg),
		Sessions:     sessions,
		Validator:    validate.New(),
		AuthStore:    store,
	}

	h := handler.New(
		cfg,
		func() echo.Routes { return e.Router().Routes() },
		&service.Services{User: service.NewUserService(store)},
		container.Validator,
	)
	availabilityH := handler.NewAvailability(func(context.Context) error { return nil }, nil, "")
	authH := auth.NewHandler(nil, auth.HandlerConfig{
		Sessions:        &sessionManagerAdapter{mgr: sessions},
		Forms:           &formParserAdapter{v: container.Validator},
		Flash:           &flashAdapter{},
		Redirect:        &redirectAdapter{},
		Pages:           authview.NewRenderer(),
		CSRFField:       cfg.App.Security.CSRF.FormField,
		SigninPath:      "/signin",
		SignupPath:      "/signup",
		VerifyPath:      "/verify",
		VerifyURL:       "http://localhost:3000/verify",
		EmailChangePath: "/account/email",
		HomePath:        "/",
		RequestFn: func(ec *echo.Context) auth.Request {
			req := view.NewRequest(ec, cfg)
			return auth.Request{
				CSRFToken: req.CSRFToken,
				PageFn:    req.Page,
			}
		},
	})

	if err := RegisterRoutes(container, h, availabilityH, authH); err != nil {
		t.Fatalf("RegisterRoutes() error = %v", err)
	}

	// makeCookie returns a session cookie for the given user ID.
	makeCookie := func(t *testing.T, userID string) *http.Cookie {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		state, err := sessions.Load(req)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if err := state.Put("auth.user_id", userID); err != nil {
			t.Fatalf("Put() error = %v", err)
		}
		if err := sessions.Commit(rec, req, state); err != nil {
			t.Fatalf("Commit() error = %v", err)
		}
		cookies := rec.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("Commit() produced no cookies")
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
