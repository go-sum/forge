package view

import (
	"net/http"
	"net/url"

	auth "github.com/go-sum/auth"
	"github.com/go-sum/componentry/patterns/flash"
	htmx "github.com/go-sum/componentry/patterns/htmx"
	render "github.com/go-sum/componentry/render/echo"
	"github.com/go-sum/componentry/patterns/font"
	"github.com/go-sum/componentry/ui/feedback"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/view/layout"
	"github.com/go-sum/server/route"

	"github.com/labstack/echo/v5"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Request collects request-scoped presentation data needed by pages and layout.
// It keeps handlers from reassembling the same shell state on every render.
type Request struct {
	CurrentPath     string
	CSRFToken       string
	CSRFFieldName   string
	CSRFHeaderName  string
	FaviconPath     string
	Description     string
	MetaKeywords    []string
	OGImage         string
	IsAuthenticated bool
	UserID          string
	UserRole        string
	UserName        string
	Flash           []flash.Message
	PasskeyEnabled  bool
	NavConfig       config.NavConfig
	FontConfig      font.Config
	CopyrightYear   int
	AppVersion      string
	HTMX            htmx.Request
	Routes          echo.Routes
}

// NewRequest builds request-scoped presentation state from the Echo context.
func NewRequest(c *echo.Context, cfg *config.Config) Request {
	req := Request{
		CurrentPath:    c.Request().URL.Path,
		CSRFFieldName:  cfg.Security.CSRF.FormField,
		CSRFHeaderName: cfg.Security.CSRF.HeaderName,
		FaviconPath:    cfg.Site.FaviconPath,
		Description:    cfg.Site.Description,
		MetaKeywords:   cfg.Site.MetaKeywords,
		OGImage:        cfg.Site.OGImage,
		FontConfig:     cfg.Site.Fonts,
		CopyrightYear:  cfg.Site.CopyrightYear,
		AppVersion:     cfg.App.Version,
		HTMX:           htmx.NewRequest(c.Request()),
		Routes:         c.Echo().Router().Routes(),
	}
	req.PasskeyEnabled = cfg.Service.Auth.Methods.Passkey.Enabled
	if !req.PasskeyEnabled {
		passkeyPath, ok := route.SafeReverse(req.Routes, "passkey.list")
		if !ok || passkeyPath == "" {
			passkeyPath = "/account/passkeys"
		}
		req.NavConfig = removeNavHref(cfg.Nav, passkeyPath)
	} else {
		req.NavConfig = cfg.Nav
	}

	if userID := auth.UserID(c); userID != "" {
		req.UserID = userID
		req.IsAuthenticated = true
	}
	if userRole := auth.UserRole(c); userRole != "" {
		req.UserRole = userRole
	}
	if name := auth.DisplayName(c); name != "" {
		req.UserName = name
	}
	if csrf, ok := c.Get(cfg.Security.CSRF.ContextKey).(string); ok && csrf != "" {
		req.CSRFToken = csrf
	}
	if flashMsgs, err := flash.GetAll(c.Request(), c.Response()); err == nil {
		req.Flash = flashMsgs
	}

	return req
}

func (r Request) Path(name string, pathValues ...any) string {
	return route.Reverse(r.Routes, name, pathValues...)
}

// safePath resolves a named route, returning fallback when routes are
// unavailable (e.g. in tests that don't register the full route table).
func (r Request) safePath(name, fallback string) string {
	if r.Routes == nil {
		return fallback
	}
	if path, ok := route.SafeReverse(r.Routes, name); ok {
		return path
	}
	return fallback
}

func (r Request) PathWithQuery(name string, query url.Values, pathValues ...any) string {
	return route.ReverseWithQuery(r.Routes, name, query, pathValues...)
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
		FaviconPath:     r.FaviconPath,
		Description:     r.Description,
		MetaKeywords:    r.MetaKeywords,
		OGImage:         r.OGImage,
		CSRFFieldName:   r.CSRFFieldName,
		CSRFHeaderName:  r.CSRFHeaderName,
		CurrentPath:     r.CurrentPath,
		CSRFToken:       r.CSRFToken,
		IsAuthenticated: r.IsAuthenticated,
		UserName:        r.UserName,
		Flash:           r.Flash,
		NavConfig:       r.NavConfig,
		FontConfig:      r.FontConfig,
		SignoutPath:      r.safePath("profile.signout.post", ""),
		CopyrightYear:   r.CopyrightYear,
		AppVersion:      r.AppVersion,
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

// removeNavHref returns a deep copy of nav with all items matching href removed.
func removeNavHref(nav config.NavConfig, href string) config.NavConfig {
	result := nav
	sections := make([]config.NavSection, len(nav.Sections))
	for i, section := range nav.Sections {
		s := section
		s.Items = filterItemsByHref(section.Items, href)
		sections[i] = s
	}
	result.Sections = sections
	return result
}

func filterItemsByHref(items []config.NavItem, href string) []config.NavItem {
	result := make([]config.NavItem, 0, len(items))
	for _, item := range items {
		if item.Href == href {
			continue
		}
		if len(item.Items) > 0 {
			item.Items = filterItemsByHref(item.Items, href)
		}
		result = append(result, item)
	}
	return result
}
