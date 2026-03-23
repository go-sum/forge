package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sum/server/apperr"
	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
)

// newTestEcho returns an Echo instance with the minimum routes registered so
// that writeErrorPage can resolve named routes (e.g. "home.show") without panicking.
func newTestEcho() *echo.Echo {
	e := echo.New()
	noOp := func(c *echo.Context) error { return c.NoContent(http.StatusOK) }
	route.Add(e, echo.Route{Method: http.MethodGet, Path: "/", Name: "home.show", Handler: noOp})
	return e
}

func TestErrorHandlerWritesProblemJSON(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	rec.Header().Set(echo.HeaderXRequestID, "req-123")

	NewErrorHandler(ErrorHandlerConfig{})(c, apperr.NotFound("The requested user could not be found."))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get(echo.HeaderContentType); !strings.Contains(got, problemContentType) {
		t.Fatalf("content-type = %q", got)
	}

	var pd problemDetails
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Fatalf("unmarshal problem details: %v", err)
	}
	if pd.Status != http.StatusNotFound || pd.Code != string(apperr.CodeNotFound) {
		t.Fatalf("problem details = %#v", pd)
	}
	if pd.Detail != "The requested user could not be found." {
		t.Fatalf("detail = %q", pd.Detail)
	}
	if pd.RequestID != "req-123" {
		t.Fatalf("request_id = %q", pd.RequestID)
	}
}

func TestErrorHandlerRendersProductionHTMLPage(t *testing.T) {
	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMETextHTML)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	NewErrorHandler(ErrorHandlerConfig{Debug: false})(c, apperr.Internal(errors.New("database timeout")))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Something went wrong on our side. Please try again.") {
		t.Fatalf("body = %q", body)
	}
	if strings.Contains(body, "database timeout") {
		t.Fatalf("production body leaked internal error: %q", body)
	}
}

func TestErrorHandlerRendersDevelopmentDetail(t *testing.T) {
	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMETextHTML)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	NewErrorHandler(ErrorHandlerConfig{Debug: true})(c, apperr.Internal(errors.New("database timeout")))

	body := rec.Body.String()
	if !strings.Contains(body, "Technical Detail") {
		t.Fatalf("body = %q", body)
	}
	if !strings.Contains(body, "database timeout") {
		t.Fatalf("development body missing internal error: %q", body)
	}
}

func TestErrorHandlerWritesHTMXToast(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/123/edit", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	NewErrorHandler(ErrorHandlerConfig{})(c, apperr.NotFound("The requested user could not be found."))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `hx-swap-oob="beforeend:#toast-container"`) {
		t.Fatalf("body = %q", body)
	}
	if !strings.Contains(body, "The requested user could not be found.") {
		t.Fatalf("body = %q", body)
	}
}

func TestErrorHandlerTreatsBoostedHTMXAsFullPage(t *testing.T) {
	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Boosted", "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	NewErrorHandler(ErrorHandlerConfig{})(c, apperr.NotFound("The requested user could not be found."))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, `hx-swap-oob="beforeend:#toast-container"`) || !strings.Contains(body, "<html") {
		t.Fatalf("body = %q", body)
	}
}

func TestErrorHandlerReturnsWithoutWritingWhenCommitted(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Response().WriteHeader(http.StatusAccepted)

	NewErrorHandler(ErrorHandlerConfig{})(c, apperr.Internal(errors.New("boom")))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestErrorHandlerIgnoresContextCanceled(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	NewErrorHandler(ErrorHandlerConfig{})(c, context.Canceled)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestClassifyMapsHTTPStatusCodeError(t *testing.T) {
	appErr := classify(echo.ErrNotFound)
	if appErr.Status != http.StatusNotFound {
		t.Fatalf("status = %d", appErr.Status)
	}
	if appErr.Code != apperr.CodeNotFound {
		t.Fatalf("code = %q", appErr.Code)
	}
}
