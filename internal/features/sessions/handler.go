package sessions

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	auth "github.com/go-sum/auth"
	"github.com/go-sum/componentry/patterns/flash"
	"github.com/go-sum/componentry/patterns/redirect"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/page"
	"github.com/go-sum/server/apperr"
	"github.com/go-sum/session"
	"github.com/labstack/echo/v5"
)

// Handler serves the session management UI.
type Handler struct {
	cfg *config.Config
	mgr session.MultiManager // nil if cookie-only mode
}

// NewHandler constructs a Handler with the given config and optional MultiManager.
func NewHandler(cfg *config.Config, mgr session.MultiManager) *Handler {
	return &Handler{cfg: cfg, mgr: mgr}
}

// List renders the active sessions page.
func (h *Handler) List(c *echo.Context) error {
	ctx := c.Request().Context()
	req := view.NewRequest(c, h.cfg)
	userID := auth.UserID(c)

	if h.mgr == nil {
		data := page.SessionListData{CookieMode: true}
		return view.Render(c, req, page.SessionListPage(req, data), page.SessionListRegion(req, data))
	}

	state, err := h.mgr.Load(c.Request())
	if err != nil {
		return apperr.Unavailable("Unable to load session data.", err)
	}
	currentID := state.ID()

	metas, err := h.mgr.ListUserSessions(ctx, userID)
	if err != nil {
		return apperr.Unavailable("Unable to load your sessions right now.", err)
	}

	entries := make([]page.SessionEntry, len(metas))
	for i, m := range metas {
		entries[i] = page.SessionEntry{
			SessionMeta: m,
			IsCurrent:   m.SessionID == currentID,
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.Before(entries[j].CreatedAt)
	})

	data := page.SessionListData{Sessions: entries}
	return view.Render(c, req, page.SessionListPage(req, data), page.SessionListRegion(req, data))
}

// Revoke terminates a specific session by ID.
func (h *Handler) Revoke(c *echo.Context) error {
	ctx := c.Request().Context()
	req := view.NewRequest(c, h.cfg)
	sessionID := c.Param("id")
	if sessionID == "" {
		return apperr.BadRequest("Session ID is required.")
	}
	userID := auth.UserID(c)

	if h.mgr == nil {
		return apperr.Unavailable("Session management is not available.", nil)
	}

	state, _ := h.mgr.Load(c.Request())
	if state != nil && state.ID() == sessionID {
		if ferr := flash.Error(c.Response(), "Use sign out to end your current session."); ferr != nil {
			slog.ErrorContext(ctx, "flash error", "error", ferr)
		}
		return redirect.New(c.Response(), c.Request()).To(req.Path("profile.session.list")).Go()
	}

	if err := h.mgr.DestroySession(ctx, sessionID, userID); err != nil {
		if errors.Is(err, session.ErrSessionNotOwned) {
			return apperr.Forbidden("That session does not belong to your account.")
		}
		slog.ErrorContext(ctx, "failed to revoke session", "session_id", sessionID, "error", err)
		return apperr.Unavailable("Unable to revoke that session right now.", err)
	}

	return c.String(http.StatusOK, "")
}

// RevokeAll terminates all sessions for the current user except the current one.
func (h *Handler) RevokeAll(c *echo.Context) error {
	ctx := c.Request().Context()
	req := view.NewRequest(c, h.cfg)
	userID := auth.UserID(c)

	if h.mgr == nil {
		return apperr.Unavailable("Session management is not available.", nil)
	}

	state, _ := h.mgr.Load(c.Request())
	var currentID string
	if state != nil {
		currentID = state.ID()
	}

	metas, err := h.mgr.ListUserSessions(ctx, userID)
	if err != nil {
		return apperr.Unavailable("Unable to load sessions right now.", err)
	}

	var failCount int
	for _, m := range metas {
		if m.SessionID == currentID {
			continue
		}
		if err := h.mgr.DestroySession(ctx, m.SessionID, userID); err != nil {
			slog.ErrorContext(ctx, "failed to destroy session during revoke-all",
				"session_id", m.SessionID, "error", err)
			failCount++
		}
	}

	var flashErr error
	if failCount > 0 {
		msg := fmt.Sprintf("%d session(s) could not be revoked. Try again.", failCount)
		flashErr = flash.Error(c.Response(), msg)
	} else {
		flashErr = flash.Success(c.Response(), "All other sessions have been signed out.")
	}
	if flashErr != nil {
		slog.ErrorContext(ctx, "flash error in revoke-all", "error", flashErr)
	}
	return redirect.New(c.Response(), c.Request()).To(req.Path("profile.session.list")).Go()
}
