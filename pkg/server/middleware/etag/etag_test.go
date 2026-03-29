package etag_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-sum/server/cache"
	"github.com/go-sum/server/middleware/etag"
	"github.com/labstack/echo/v5"
)

// newGETContext creates an Echo GET context backed by a test response recorder.
func newGETContext(path string) (*echo.Echo, *echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return e, c, rec
}

// handlerWriting returns a handler that writes the given body with 200 OK.
func handlerWriting(body string) echo.HandlerFunc {
	return func(c *echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/plain")
		c.Response().WriteHeader(http.StatusOK)
		_, err := c.Response().Write([]byte(body))
		return err
	}
}

func TestMiddleware_FirstGET_200WithETag(t *testing.T) {
	_, c, rec := newGETContext("/")

	err := etag.Middleware()(handlerWriting("hello"))(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "hello" {
		t.Errorf("body = %q, want %q", got, "hello")
	}
	etagVal := rec.Header().Get("ETag")
	if etagVal == "" {
		t.Error("ETag header not set on first request")
	}
	want := cache.WeakETag([]byte("hello"))
	if etagVal != want {
		t.Errorf("ETag = %q, want %q", etagVal, want)
	}
}

func TestMiddleware_MatchingIfNoneMatch_304(t *testing.T) {
	e := echo.New()
	etagVal := cache.WeakETag([]byte("hello"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", etagVal)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := etag.Middleware()(handlerWriting("hello"))(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotModified {
		t.Errorf("status = %d, want 304", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body should be empty for 304, got %q", rec.Body.String())
	}
	if got := rec.Header().Get("ETag"); got == "" {
		t.Error("ETag header should be present on 304 response")
	}
}

func TestMiddleware_WildcardIfNoneMatch_304(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", "*")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := etag.Middleware()(handlerWriting("hello"))(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotModified {
		t.Errorf("status = %d, want 304", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body should be empty for 304, got %q", rec.Body.String())
	}
}

func TestMiddleware_StaleETag_200(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", `W/"stale-tag-does-not-match"`)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := etag.Middleware()(handlerWriting("hello"))(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "hello" {
		t.Errorf("body = %q, want %q", got, "hello")
	}
	if got := rec.Header().Get("ETag"); got == "" {
		t.Error("ETag header not set on 200 response with stale ETag")
	}
}

func TestMiddleware_Skipper_NoETag(t *testing.T) {
	_, c, rec := newGETContext("/")

	cfg := etag.Config{
		Skipper: func(*echo.Context) bool { return true },
	}
	err := etag.NewWithConfig(cfg)(handlerWriting("hello"))(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("ETag"); got != "" {
		t.Errorf("ETag should not be set when Skipper returns true, got %q", got)
	}
}

func TestMiddleware_POST_Passthrough(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := etag.Middleware()(handlerWriting("hello"))(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("ETag"); got != "" {
		t.Errorf("ETag should not be set for POST, got %q", got)
	}
}

func TestMiddleware_HandlerError_Propagates(t *testing.T) {
	_, c, rec := newGETContext("/")

	sentinelErr := errors.New("handler error")
	handler := func(c *echo.Context) error {
		return sentinelErr
	}

	err := etag.Middleware()(handler)(c)
	if !errors.Is(err, sentinelErr) {
		t.Errorf("error = %v, want %v", err, sentinelErr)
	}
	// Nothing should have been written to the real recorder.
	if rec.Body.Len() != 0 {
		t.Errorf("body should be empty when handler errors, got %q", rec.Body.String())
	}
	// Committed should remain false so Echo's error handler can write.
	// Use UnwrapResponse to access the *echo.Response fields — c.Response()
	// returns http.ResponseWriter (the interface), not *echo.Response directly.
	if resp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil && resp.Committed {
		t.Error("response should not be committed when handler returns error without writing")
	}
}
