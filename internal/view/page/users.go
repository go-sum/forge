package page

import (
	"starter/internal/model"
	"starter/internal/routes"
	"starter/internal/view"
	"starter/internal/view/partial/userpartial"
	componenthtmx "starter/pkg/components/patterns/htmx"
	"starter/pkg/components/patterns/pager"
	"starter/pkg/components/ui/core"
	uidata "starter/pkg/components/ui/data"

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
		h.H1(h.Class("text-2xl font-bold mb-4"), g.Text("Users")),
		UserListRegion(data),
	)
}

// UserListRegion renders the HTMX-replaceable table + pagination region.
func UserListRegion(data UserListData) g.Node {
	return h.Div(
		h.ID("users-list-region"),
		h.Class("space-y-4"),
		userTable(data.Users),
		pagination(data.Pager),
	)
}

func userTable(users []model.User) g.Node {
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
		h.TBody(
			h.ID("users-table"),
			h.Class("[&_tr:last-child]:border-0"),
			g.Group(rows),
		),
	)
}

func pagination(p pager.Pager) g.Node {
	if p.TotalPages <= 1 {
		return g.Text("")
	}
	return h.Div(
		h.Class("flex justify-center gap-2 mt-4"),
		g.If(!p.IsFirst(),
			core.Button(core.ButtonProps{
				Label:   "← Previous",
				Href:    routes.UserListPage(p.PrevPage()),
				Variant: core.VariantGhost,
				Size:    core.SizeSm,
				Extra: componenthtmx.PaginatedTableLink(componenthtmx.PaginatedTableProps{
					Path:    routes.Users,
					Page:    p.PrevPage(),
					Target:  "#users-list-region",
					Swap:    componenthtmx.SwapOuterHTML,
					PushURL: true,
				}),
			}),
		),
		h.Span(
			h.Class("text-sm self-center text-muted-foreground"),
			g.Textf("Page %d of %d", p.Page, p.TotalPages),
		),
		g.If(!p.IsLast(),
			core.Button(core.ButtonProps{
				Label:   "Next →",
				Href:    routes.UserListPage(p.NextPage()),
				Variant: core.VariantGhost,
				Size:    core.SizeSm,
				Extra: componenthtmx.PaginatedTableLink(componenthtmx.PaginatedTableProps{
					Path:    routes.Users,
					Page:    p.NextPage(),
					Target:  "#users-list-region",
					Swap:    componenthtmx.SwapOuterHTML,
					PushURL: true,
				}),
			}),
		),
	)
}
