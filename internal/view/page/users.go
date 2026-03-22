// Example: safe to delete as a unit.
package page

import (
	"fmt"
	"github.com/y-goweb/foundry/internal/model"
	"github.com/y-goweb/foundry/internal/routes"
	"github.com/y-goweb/foundry/internal/view"
	"github.com/y-goweb/foundry/internal/view/partial/userpartial"
	uipagination "github.com/y-goweb/componentry/interactive/pagination"
	componenthtmx "github.com/y-goweb/componentry/patterns/htmx"
	"github.com/y-goweb/componentry/patterns/pager"
	"github.com/y-goweb/componentry/ui/core"
	uidata "github.com/y-goweb/componentry/ui/data"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// UserListData configures the user management table and pagination region.
type UserListData struct {
	Users []model.User
	Pager pager.Pager
}

// UserListPage renders the full user table inside the base layout.
func UserListPage(req view.Request, data UserListData) g.Node {
	return req.Page(
		"Users",
		h.Div(
			h.Class("space-y-6"),
			usersPageHeader(data.Pager.TotalItems),
			UserListRegion(data),
		),
	)
}

// UserListRegion renders the HTMX-replaceable table + pagination region.
func UserListRegion(data UserListData) g.Node {
	return h.Div(
		h.ID("users-list-region"),
		h.Class("space-y-4"),
		userTable(data.Users),
		usersLoadingIndicator(),
		pagination(data.Pager),
	)
}

func usersPageHeader(total int) g.Node {
	return h.Div(
		h.Class("flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between"),
		h.Div(
			h.Class("space-y-2"),
			h.H1(h.Class("text-2xl font-bold"), g.Text("Users")),
			h.P(
				h.Class("max-w-2xl text-sm text-muted-foreground"),
				g.Text("Manage account records with inline edits and lightweight HTMX updates."),
			),
		),
		h.P(
			h.Class("text-sm text-muted-foreground"),
			g.Text(userCountLabel(total)),
		),
	)
}

func userTable(users []model.User) g.Node {
	if len(users) == 0 {
		return emptyUsersState()
	}

	rows := make([]g.Node, len(users))
	for i, u := range users {
		rows[i] = userpartial.UserRow(userpartial.UserRowProps{
			User: u,
		})
	}

	return uidata.Table.Root(
		uidata.Table.Header(
			uidata.Table.Row(uidata.RowProps{},
				uidata.Table.Head(g.Text("Name")),
				uidata.Table.Head(g.Text("Email")),
				uidata.Table.Head(g.Text("Role")),
				uidata.Table.Head(g.Text("Created")),
				uidata.Table.Head(g.Text("Actions")),
			),
		),
		uidata.Table.Body(uidata.BodyProps{
			ID: "users-table",
		},
			g.Group(rows),
		),
	)
}

func pagination(p pager.Pager) g.Node {
	if p.TotalPages <= 1 {
		return g.Text("")
	}
	items := []g.Node{
		uipagination.Item(uipagination.Previous(routes.UserListPage(p.PrevPage()), p.IsFirst(), paginatedLinkAttrs(p.PrevPage())...)),
	}
	for _, pageNumber := range paginationSequence(p) {
		if pageNumber == 0 {
			items = append(items, uipagination.Item(uipagination.Ellipsis()))
			continue
		}
		items = append(items, uipagination.Item(uipagination.Link(
			routes.UserListPage(pageNumber),
			pageNumber == p.Page,
			append(paginatedLinkAttrs(pageNumber), g.Textf("%d", pageNumber))...,
		)))
	}
	items = append(items,
		uipagination.Item(uipagination.Next(routes.UserListPage(p.NextPage()), p.IsLast(), paginatedLinkAttrs(p.NextPage())...)),
	)

	return h.Div(
		h.Class("space-y-2"),
		uipagination.Root(
			uipagination.Content(g.Group(items)),
		),
		h.P(
			h.Class("text-center text-sm text-muted-foreground"),
			g.Textf("Page %d of %d", p.Page, p.TotalPages),
		),
	)
}

func emptyUsersState() g.Node {
	return uidata.Card.Root(
		uidata.Card.Content(
			h.Div(
				h.Class("flex flex-col items-center justify-center gap-2 py-10 text-center"),
				h.H2(h.Class("text-lg font-semibold"), g.Text("No users yet")),
				h.P(
					h.Class("max-w-md text-sm text-muted-foreground"),
					g.Text("User accounts will appear here once people register. This table also handles inline edits and row-level actions when data is present."),
				),
			),
		),
	)
}

func usersLoadingIndicator() g.Node {
	return h.Div(
		h.ID("users-loading"),
		h.Class("htmx-indicator flex items-center justify-center gap-3 rounded-md border border-dashed border-border bg-muted/20 px-4 py-3 text-sm text-muted-foreground"),
		core.Skeleton(h.Class("h-2 w-16")),
		h.Span(g.Text("Updating users...")),
	)
}

func userCountLabel(total int) string {
	if total == 1 {
		return "1 user"
	}
	return fmt.Sprintf("%d users", total)
}

func paginatedLinkAttrs(page int) []g.Node {
	return componenthtmx.PaginatedTableLink(componenthtmx.PaginatedTableProps{
		Path:      routes.Users,
		Page:      page,
		Target:    "#users-list-region",
		Swap:      componenthtmx.SwapOuterHTML,
		PushURL:   true,
		Indicator: "#users-loading",
	})
}

func paginationSequence(p pager.Pager) []int {
	if p.TotalPages <= 0 {
		return nil
	}
	if p.TotalPages <= 5 {
		pages := make([]int, 0, p.TotalPages)
		for i := 1; i <= p.TotalPages; i++ {
			pages = append(pages, i)
		}
		return pages
	}

	pages := []int{1}
	start := max(2, p.Page-1)
	end := min(p.TotalPages-1, p.Page+1)
	if start > 2 {
		pages = append(pages, 0)
	}
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	if end < p.TotalPages-1 {
		pages = append(pages, 0)
	}
	return append(pages, p.TotalPages)
}
