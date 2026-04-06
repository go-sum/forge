package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/security/token"
	"github.com/labstack/echo/v5"
)

func testSecurityConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Security: config.SecurityConfig{
				ExternalOrigin: "https://example.com",
				Origin: config.OriginConfig{
					Enabled:       true,
					RequireHeader: true,
				},
				FetchMetadata: config.FetchMetadataConfig{
					Enabled:                 true,
					AllowedSites:            []string{"same-origin", "same-site"},
					AllowedModes:            []string{"cors", "navigate", "same-origin"},
					FallbackWhenMissing:     true,
					RejectCrossSiteNavigate: true,
				},
				Headers: config.HeadersConfig{
					XSSProtection:         "0",
					ContentTypeNosniff:    true,
					FrameOptions:          "DENY",
					ContentSecurityPolicy: "default-src 'self'",
					HSTS: config.HSTSConfig{
						Enabled:           true,
						MaxAge:            31536000,
						IncludeSubDomains: true,
						Preload:           true,
					},
				},
			},
		},
	}
}

func TestCrossOriginGuardAllowsVerifiedUnsafeRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := CrossOriginGuard(testSecurityConfig())(func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err != nil {
		t.Fatalf("CrossOriginGuard() error = %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCrossOriginGuardWritesProblemJSON(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = NewErrorHandler(ErrorHandlerConfig{})
	e.Use(CrossOriginGuard(testSecurityConfig()))
	e.POST("/signin", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodPost, "/signin", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderContentType); !strings.Contains(got, problemContentType) {
		t.Fatalf("content-type = %q", got)
	}

	var pd problemDetails
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Fatalf("unmarshal problem details: %v", err)
	}
	if pd.Status != http.StatusForbidden {
		t.Fatalf("problem details = %#v", pd)
	}
	if !strings.Contains(pd.Detail, "origin headers") {
		t.Fatalf("detail = %q", pd.Detail)
	}
}

func TestCrossOriginGuardWritesHTMXToast(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = NewErrorHandler(ErrorHandlerConfig{})
	e.Use(CrossOriginGuard(testSecurityConfig()))
	e.POST("/signin", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodPost, "/signin", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, `hx-swap-oob="beforeend:#toast-container"`) {
		t.Fatalf("body = %q", body)
	}
}

func TestCSRFMiddlewareUsesTTLSecondsFromConfig(t *testing.T) {
	cfg := testSecurityConfig()
	cfg.App.Security.CSRF = config.CSRFConfig{
		Key:        "test-signing-key-32-bytes-padded!",
		ContextKey: "csrf",
		FormField:  "_csrf",
		HeaderName: "X-CSRF-Token",
		TokenTTL:   3600,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := CSRFMiddleware(cfg)(func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})(c)
	if err != nil {
		t.Fatalf("CSRFMiddleware() error = %v", err)
	}

	tok, ok := c.Get("csrf").(string)
	if !ok || tok == "" {
		t.Fatalf("context key %q: got %q, want non-empty token", "csrf", tok)
	}
	if err := token.Verify([]byte(cfg.App.Security.CSRF.Key), "csrf", tok); err != nil {
		t.Fatalf("token.Verify() error = %v", err)
	}
}
