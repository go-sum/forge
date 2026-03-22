package view

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/componentry/patterns/flash"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/server/ctxkeys"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

func TestNewRequestCollectsPresentationState(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users?page=2", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "#users")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(ctxkeys.UserID), "user-123")
	c.Set(string(ctxkeys.UserRole), "admin")
	c.Set(string(ctxkeys.UserDisplayName), "Alice")
	c.Set(echomw.DefaultCSRFConfig.ContextKey, "csrf-token")
	if err := flash.Success(rec, "Saved"); err != nil {
		t.Fatalf("set flash: %v", err)
	}
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}

	viewReq := NewRequest(c, config.NavConfig{})

	if viewReq.CurrentPath != "/users" {
		t.Fatalf("CurrentPath = %q", viewReq.CurrentPath)
	}
	if !viewReq.IsAuthenticated || viewReq.UserID != "user-123" {
		t.Fatalf("auth state = %#v", viewReq)
	}
	if viewReq.UserRole != "admin" {
		t.Fatalf("UserRole = %q", viewReq.UserRole)
	}
	if viewReq.UserName != "Alice" {
		t.Fatalf("UserName = %q", viewReq.UserName)
	}
	if viewReq.CSRFToken != "csrf-token" {
		t.Fatalf("CSRFToken = %q", viewReq.CSRFToken)
	}
	if !viewReq.HTMX.Enabled || viewReq.HTMX.Target != "#users" {
		t.Fatalf("HTMX = %#v", viewReq.HTMX)
	}
	if len(viewReq.Flash) != 1 || viewReq.Flash[0].Text != "Saved" {
		t.Fatalf("Flash = %#v", viewReq.Flash)
	}
}
