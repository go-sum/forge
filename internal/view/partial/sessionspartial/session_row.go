package sessionspartial

import (
	htmx "github.com/go-sum/componentry/patterns/htmx"
	"github.com/go-sum/componentry/ui/core"
	uidata "github.com/go-sum/componentry/ui/data"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/session"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// SessionEntry pairs session metadata with a flag for the current session.
type SessionEntry struct {
	session.SessionMeta
	IsCurrent bool
}

// SessionRow renders a single session row in the sessions management table.
func SessionRow(req view.Request, entry SessionEntry) g.Node {
	sid := entry.SessionID
	shortID := sid
	if len(sid) > 8 {
		shortID = sid[:8]
	}

	rowAttrs := []g.Node{
		h.ID("session-" + shortID),
	}
	if entry.IsCurrent {
		rowAttrs = append(rowAttrs, h.Class("bg-muted/30"))
	}

	ip := entry.IPAddress
	if ip == "" {
		ip = "—"
	}

	return uidata.Table.Row(uidata.RowProps{}, append(rowAttrs,
		uidata.Table.Cell(methodCell(entry)),
		uidata.Table.Cell(g.Text(ip)),
		uidata.Table.Cell(g.Text(entry.LastActiveAt.Format("2006-01-02 15:04"))),
		uidata.Table.Cell(g.Text(entry.CreatedAt.Format("2006-01-02 15:04"))),
		uidata.Table.Cell(actionsCell(req, entry)),
	)...)
}

func methodCell(entry SessionEntry) g.Node {
	return h.Div(
		h.Class("flex items-center gap-2"),
		h.Span(g.Text(friendlyMethod(entry.AuthMethod))),
		g.If(entry.IsCurrent,
			core.Badge(core.BadgeProps{
				Variant:  core.BadgeSecondary,
				Children: []g.Node{g.Text("Current")},
			}),
		),
	)
}

func actionsCell(req view.Request, entry SessionEntry) g.Node {
	if entry.IsCurrent {
		return h.Span(h.Class("text-xs text-muted-foreground"), g.Text("Current session"))
	}
	return core.Button(core.ButtonProps{
		Label:   "Sign out",
		Variant: core.VariantDestructiveGhost,
		Size:    core.SizeSm,
		Type:    "button",
		Extra: htmx.Attrs(htmx.AttrsProps{
			Delete:    req.Path("session.revoke", entry.SessionID),
			Confirm:   "Sign out this session?",
			Target:    "closest tr",
			Swap:      "outerHTML swap:300ms",
			Indicator: "#sessions-loading",
		}),
	})
}

func friendlyMethod(method string) string {
	switch method {
	case "email_totp":
		return "Email + Code"
	case "passkey":
		return "Passkey"
	default:
		return "Unknown"
	}
}
