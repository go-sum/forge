package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/assets/publish"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/features/availability"
	appserver "github.com/go-sum/forge/internal/server"
	"github.com/labstack/echo/v5"
)

func TestRegisterStartupRoutesServesUnavailableAndHealth(t *testing.T) {
	if err := publish.Init("public", "/public"); err != nil {
		t.Fatalf("publish.Init() error = %v", err)
	}

	e := echo.New()
	cfg := &config.Config{
		Security: config.SecurityConfig{
			CSRF: config.CSRFConfig{ContextKey: "csrf"},
		},
		Nav: config.NavConfig{
			Brand: config.NavbarBrand{Label: "Starter", Href: "/"},
		},
	}
	e.HTTPErrorHandler = appserver.NewErrorHandler(appserver.ErrorHandlerConfig{
		Config: cfg,
	})

	runtime := &Runtime{
		Config:       cfg,
		Web:          e,
		PublicPrefix: "/public",
		PublicDir:    "public",
	}
	startupH := availability.NewHandler(
		func(context.Context) error { return errors.New("database verify: missing required relations users") },
		errors.New("database verify: missing required relations users"),
		"",
	)

	if err := RegisterStartupRoutes(runtime, startupH); err != nil {
		t.Fatalf("RegisterStartupRoutes() error = %v", err)
	}

	reqPage := httptest.NewRequest(http.MethodGet, "/", nil)
	recPage := httptest.NewRecorder()
	e.ServeHTTP(recPage, reqPage)
	if recPage.Code != http.StatusServiceUnavailable {
		t.Fatalf("page status = %d, want %d", recPage.Code, http.StatusServiceUnavailable)
	}

	reqHealth := httptest.NewRequest(http.MethodGet, "/health", nil)
	recHealth := httptest.NewRecorder()
	e.ServeHTTP(recHealth, reqHealth)
	if recHealth.Code != http.StatusServiceUnavailable {
		t.Fatalf("health status = %d, want %d", recHealth.Code, http.StatusServiceUnavailable)
	}
	if recHealth.Body.String() != "{\"status\":\"error\"}\n" {
		t.Fatalf("health body = %q, want %q", recHealth.Body.String(), "{\"status\":\"error\"}\n")
	}
}
