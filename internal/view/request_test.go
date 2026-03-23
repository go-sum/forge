package view

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/componentry/patterns/flash"
	"github.com/go-sum/forge/config"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

var testKeys = config.ContextKeysConfig{
	UserID:      "user_id",
	UserRole:    "user_role",
	DisplayName: "user_display_name",
	CSRF:        "csrf",
}

func TestNewRequestCollectsPresentationState(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users?page=2", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "#users")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(testKeys.UserID, "user-123")
	c.Set(testKeys.UserRole, "admin")
	c.Set(testKeys.DisplayName, "Alice")
	c.Set(echomw.DefaultCSRFConfig.ContextKey, "csrf-token")
	if err := flash.Success(rec, "Saved"); err != nil {
		t.Fatalf("set flash: %v", err)
	}
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}

	viewReq := NewRequest(c, &config.Config{
		Keys:   testKeys,
		Server: config.ServerConfig{CSRFCookieName: "_csrf"},
		Site:   config.SiteConfig{FaviconPath: "/public/img/favicon.ico"},
	})

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
	if viewReq.CSRFFieldName != "_csrf" {
		t.Fatalf("CSRFFieldName = %q", viewReq.CSRFFieldName)
	}
	if viewReq.FaviconPath != "/public/img/favicon.ico" {
		t.Fatalf("FaviconPath = %q", viewReq.FaviconPath)
	}
	if !viewReq.HTMX.Enabled || viewReq.HTMX.Target != "#users" {
		t.Fatalf("HTMX = %#v", viewReq.HTMX)
	}
	if len(viewReq.Flash) != 1 || viewReq.Flash[0].Text != "Saved" {
		t.Fatalf("Flash = %#v", viewReq.Flash)
	}
}
