package view

import (
	"net/http"
	"net/url"

	"github.com/go-sum/componentry/patterns/flash"
	componenthtmx "github.com/go-sum/componentry/patterns/htmx"
	render "github.com/go-sum/componentry/render/echo"
	"github.com/go-sum/componentry/ui/feedback"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/view/layout"
	"github.com/go-sum/server/ctxkeys"
	serverroute "github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Request collects request-scoped presentation data needed by pages and layout.
// It keeps handlers from reassembling the same shell state on every render.
type Request struct {
	CurrentPath     string
	CSRFToken       string
	IsAuthenticated bool
	UserID          string
	UserRole        string
	UserName        string
	Flash           []flash.Message
	NavConfig       config.NavConfig
	HTMX            componenthtmx.Request
	Routes          echo.Routes
}

// NewRequest builds request-scoped presentation state from the Echo context.
func NewRequest(c *echo.Context, navConfig config.NavConfig) Request {
	req := Request{
		CurrentPath: c.Request().URL.Path,
		NavConfig:   navConfig,
		HTMX:        componenthtmx.NewRequest(c.Request()),
		Routes:      c.Echo().Router().Routes(),
	}

	if userID, _ := c.Get(string(ctxkeys.UserID)).(string); userID != "" {
		req.UserID = userID
		req.IsAuthenticated = true
	}
	if userRole, _ := c.Get(string(ctxkeys.UserRole)).(string); userRole != "" {
		req.UserRole = userRole
	}
	if name, _ := c.Get(string(ctxkeys.UserDisplayName)).(string); name != "" {
		req.UserName = name
	}
	if csrf, _ := c.Get(echomw.DefaultCSRFConfig.ContextKey).(string); csrf != "" {
		req.CSRFToken = csrf
	}
	if flashMsgs, err := flash.GetAll(c.Request(), c.Response()); err == nil {
		req.Flash = flashMsgs
	}

	return req
}

func (r Request) Path(name string, pathValues ...any) string {
	return serverroute.Reverse(r.Routes, name, pathValues...)
}

func (r Request) PathWithQuery(name string, query url.Values, pathValues ...any) string {
	return serverroute.ReverseWithQuery(r.Routes, name, query, pathValues...)
}

// FormError renders a destructive alert listing validation messages.
// Returns an empty text node when messages is empty so callers need no nil check.
func FormError(messages []string) g.Node {
	if len(messages) == 0 {
		return g.Text("")
	}
	items := make([]g.Node, len(messages))
	for i, msg := range messages {
		items[i] = h.Li(g.Text(msg))
	}
	return feedback.Alert.Root(
		feedback.AlertProps{Variant: feedback.AlertDestructive},
		feedback.Alert.Description(
			h.Ul(h.Class("list-disc space-y-1 pl-4"), g.Group(items)),
		),
	)
}

// IsPartial reports whether the request should receive a fragment response.
func (r Request) IsPartial() bool {
	return r.HTMX.IsPartial()
}

// LayoutProps builds the shared layout props for a full-page render.
func (r Request) LayoutProps(title string, children ...g.Node) layout.Props {
	return layout.Props{
		Title:           title,
		CurrentPath:     r.CurrentPath,
		CSRFToken:       r.CSRFToken,
		IsAuthenticated: r.IsAuthenticated,
		UserName:        r.UserName,
		Flash:           r.Flash,
		NavConfig:       r.NavConfig,
		Children:        children,
	}
}

// Page wraps children with the shared application layout.
func (r Request) Page(title string, children ...g.Node) g.Node {
	return layout.Page(r.LayoutProps(title, children...))
}

// Render chooses the correct response mode for the request. HTMX partial
// requests receive partial, all others receive full. If partial is nil, full is used.
func Render(c *echo.Context, req Request, full, partial g.Node) error {
	return RenderWithStatus(c, req, http.StatusOK, full, partial)
}

// RenderWithStatus is Render with an explicit HTTP status code.
func RenderWithStatus(c *echo.Context, req Request, status int, full, partial g.Node) error {
	if partial == nil {
		partial = full
	}
	if req.IsPartial() {
		return render.FragmentWithStatus(c, status, partial)
	}
	return render.ComponentWithStatus(c, status, full)
}
