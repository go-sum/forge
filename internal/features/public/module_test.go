package public

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sum/componentry/patterns/flash"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/server/route"
	site "github.com/go-sum/site"
	sitehandlers "github.com/go-sum/site/handlers"
	"github.com/labstack/echo/v5"
)

func TestModuleRendersHome(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{
		App: config.AppConfig{Security: config.SecurityConfig{CSRF: config.CSRFConfig{ContextKey: "csrf"}}},
		Nav: config.NavConfig{Brand: config.NavbarBrand{Label: "Starter", Href: "/"}},
	}
	noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/_components", Name: "components.list", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/signin", Name: "signin.get", Handler: noOp})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/users", Name: "user.list", Handler: noOp})
	siteH := sitehandlers.New(sitehandlers.Config{Origin: "http://localhost"}, func() echo.Routes { return e.Router().Routes() })
	m := NewModule(cfg, siteH)
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: m.Home})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/robots.txt", Name: "robots.show", Handler: siteH.RobotsTxt})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/sitemap.xml", Name: "sitemap.show", Handler: siteH.SitemapXML})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	flashRec := httptest.NewRecorder()
	if err := flash.Success(flashRec, "Saved"); err != nil {
		t.Fatalf("flash.Success() error = %v", err)
	}
	for _, cookie := range flashRec.Result().Cookies() {
		req.AddCookie(cookie)
	}

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Modern Web Starter") || !strings.Contains(body, "Saved") {
		t.Fatalf("body = %q", body)
	}
}

func TestModuleRegistersSiteFiles(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{}
	siteH := sitehandlers.New(sitehandlers.Config{
		Origin: "http://localhost",
		Robots: site.RobotsConfig{DefaultAllow: true, DisallowPaths: []string{"/admin"}},
	}, func() echo.Routes { return e.Router().Routes() })
	m := NewModule(cfg, siteH)
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: m.Home})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/robots.txt", Name: "robots.show", Handler: siteH.RobotsTxt})
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/sitemap.xml", Name: "sitemap.show", Handler: siteH.SitemapXML})

	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Sitemap: http://localhost/sitemap.xml") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
