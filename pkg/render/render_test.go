package render

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	g "maragu.dev/gomponents"
)

type errNode struct{}

func (errNode) Render(w io.Writer) error { return errors.New("render failed") }

func TestRenderHelpersWriteHTMLResponses(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	tests := []struct {
		name       string
		call       func(*echo.Context) error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "component",
			call:       func(c *echo.Context) error { return Component(c, g.Text("full")) },
			wantStatus: http.StatusOK,
			wantBody:   "full",
		},
		{
			name:       "component with status",
			call:       func(c *echo.Context) error { return ComponentWithStatus(c, http.StatusCreated, g.Text("created")) },
			wantStatus: http.StatusCreated,
			wantBody:   "created",
		},
		{
			name:       "fragment",
			call:       func(c *echo.Context) error { return Fragment(c, g.Text("fragment")) },
			wantStatus: http.StatusOK,
			wantBody:   "fragment",
		},
		{
			name: "fragment with status",
			call: func(c *echo.Context) error {
				return FragmentWithStatus(c, http.StatusUnprocessableEntity, g.Text("invalid"))
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   "invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if err := tc.call(c); err != nil {
				t.Fatalf("render call error = %v", err)
			}
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d", rec.Code)
			}
			if rec.Body.String() != tc.wantBody {
				t.Fatalf("body = %q", rec.Body.String())
			}
			if got := rec.Header().Get(echo.HeaderContentType); got != echo.MIMETextHTMLCharsetUTF8 {
				t.Fatalf("content-type = %q", got)
			}
		})
	}
}

func TestComponentWithStatusReturnsRenderError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := ComponentWithStatus(c, http.StatusOK, errNode{}); err == nil {
		t.Fatal("ComponentWithStatus() unexpectedly succeeded")
	}
}
