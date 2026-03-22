package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sum/forge/internal/routes"
	"github.com/go-sum/componentry/patterns/flash"
)

func TestHealthCheckReportsStatus(t *testing.T) {
	tests := []struct {
		name        string
		checkHealth func(context.Context) error
		wantStatus  int
		wantBody    string
	}{
		{
			name:        "healthy",
			checkHealth: func(context.Context) error { return nil },
			wantStatus:  http.StatusOK,
			wantBody:    `"status":"ok"`,
		},
		{
			name:        "unhealthy",
			checkHealth: func(context.Context) error { return errors.New("db down") },
			wantStatus:  http.StatusServiceUnavailable,
			wantBody:    `"status":"error"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(fakeUserService{},tc.checkHealth)
			c, rec := newRequestContext(http.MethodGet, routes.Health, nil)

			if err := h.HealthCheck(c); err != nil {
				t.Fatalf("HealthCheck() error = %v", err)
			}
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), tc.wantBody) {
				t.Fatalf("body = %q", rec.Body.String())
			}
		})
	}
}

func TestHomeRendersFlashMessages(t *testing.T) {
	h := newTestHandler(fakeUserService{},nil)
	c, rec := newRequestContext(http.MethodGet, routes.Home, nil)
	setCSRFToken(c)
	setUserID(c, testUser.ID.String())

	flashRec := httptest.NewRecorder()
	if err := flash.Success(flashRec, "Saved"); err != nil {
		t.Fatalf("flash.Success() error = %v", err)
	}
	for _, cookie := range flashRec.Result().Cookies() {
		c.Request().AddCookie(cookie)
	}

	if err := h.Home(c); err != nil {
		t.Fatalf("Home() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Modern Web Starter") || !strings.Contains(body, "Saved") {
		t.Fatalf("body = %q", body)
	}
}

func TestComponentExamplesRenders(t *testing.T) {
	h := newTestHandler(fakeUserService{},nil)
	c, rec := newRequestContext(http.MethodGet, routes.Components, nil)
	setCSRFToken(c)

	if err := h.ComponentExamples(c); err != nil {
		t.Fatalf("ComponentExamples() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Component Examples") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
