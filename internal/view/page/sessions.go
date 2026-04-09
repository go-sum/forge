package page

import (
	htmx "github.com/go-sum/componentry/patterns/htmx"
	"github.com/go-sum/componentry/ui/core"
	uidata "github.com/go-sum/componentry/ui/data"
	feedback "github.com/go-sum/componentry/ui/feedback"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/forge/internal/view/partial/sessionspartial"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// SessionEntry is an alias for the sessionspartial entry type.
type SessionEntry = sessionspartial.SessionEntry

// SessionListData configures the sessions management page.
type SessionListData struct {
	Sessions   []SessionEntry
	CookieMode bool
}

// SessionListPage renders the full sessions page inside the base layout.
func SessionListPage(req view.Request, data SessionListData) g.Node {
	return req.Page(
		"Sessions",
		h.Div(
			h.Class("space-y-6"),
			sessionsPageHeader(),
			SessionListRegion(req, data),
		),
	)
}

// SessionListRegion renders the HTMX-replaceable sessions table region.
func SessionListRegion(req view.Request, data SessionListData) g.Node {
	if data.CookieMode {
		return h.Div(
			h.ID("sessions-list-region"),
			cookieModeNotice(),
		)
	}

	hasOtherSessions := false
	for _, e := range data.Sessions {
		if !e.IsCurrent {
			hasOtherSessions = true
			break
		}
	}

	return h.Div(
		h.ID("sessions-list-region"),
		h.Class("space-y-4"),
		sessionsTable(req, data.Sessions),
		sessionsLoadingIndicator(),
		g.If(hasOtherSessions, revokeAllForm(req)),
	)
}

func sessionsPageHeader() g.Node {
	return h.Div(
		h.Class("space-y-2"),
		h.H1(h.Class("text-2xl font-bold"), g.Text("Active Sessions")),
		h.P(
			h.Class("max-w-2xl text-sm text-muted-foreground"),
			g.Text("Manage where you're currently signed in."),
		),
	)
}

func sessionsTable(req view.Request, entries []SessionEntry) g.Node {
	if len(entries) == 0 {
		return emptySessionsState()
	}
	rows := make([]g.Node, len(entries))
	for i, e := range entries {
		rows[i] = sessionspartial.SessionRow(req, e)
	}
	return uidata.Table.Root(
		uidata.Table.Header(
			uidata.Table.Row(uidata.RowProps{},
				uidata.Table.Head(g.Text("Method")),
				uidata.Table.Head(g.Text("IP Address")),
				uidata.Table.Head(g.Text("Last Active")),
				uidata.Table.Head(g.Text("Signed In")),
				uidata.Table.Head(g.Text("Actions")),
			),
		),
		uidata.Table.Body(uidata.BodyProps{ID: "sessions-table"},
			g.Group(rows),
		),
	)
}

func emptySessionsState() g.Node {
	return uidata.Card.Root(
		uidata.Card.Content(
			h.Div(
				h.Class("flex flex-col items-center justify-center gap-2 py-10 text-center"),
				h.H2(h.Class("text-lg font-semibold"), g.Text("No active sessions found.")),
			),
		),
	)
}

func sessionsLoadingIndicator() g.Node {
	return h.Div(
		h.ID("sessions-loading"),
		h.Class("htmx-indicator flex items-center justify-center gap-3 rounded-md border border-dashed border-border bg-muted/20 px-4 py-3 text-sm text-muted-foreground"),
		core.Skeleton(h.Class("h-2 w-16")),
		h.Span(g.Text("Updating sessions...")),
	)
}

func cookieModeNotice() g.Node {
	return feedback.Alert.Root(
		feedback.AlertProps{Variant: feedback.AlertDefault},
		feedback.Alert.Title(g.Text("Cookie-based sessions")),
		feedback.Alert.Description(
			g.Text("The session listing and management require server-side sessions to be enabled."),
		),
	)
}

func revokeAllForm(req view.Request) g.Node {
	return h.Div(
		h.Class("flex justify-end"),
		core.Button(core.ButtonProps{
			Label:   "Sign out everywhere else",
			Variant: core.VariantDestructiveGhost,
			Type:    "button",
			Extra: htmx.Attrs(htmx.AttrsProps{
				Delete:  req.Path("profile.session.revoke.all"),
				Confirm: "Sign out of all other sessions?",
				Target:  "#sessions-list-region",
				Swap:    htmx.SwapOuterHTML,
			}),
		}),
	)
}
