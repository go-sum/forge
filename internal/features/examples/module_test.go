package examples

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/server/route"
	"github.com/labstack/echo/v5"
)

func TestModuleRendersExamples(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{
		App: config.AppConfig{Security: config.SecurityConfig{CSRF: config.CSRFConfig{ContextKey: "csrf"}}},
		Nav: config.NavConfig{Brand: config.NavbarBrand{Label: "Starter", Href: "/"}},
	}
	m := NewModule(cfg)
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "components.list", Handler: m.Handle})

	req := httptest.NewRequest(http.MethodGet, "/_components", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Component Examples") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
