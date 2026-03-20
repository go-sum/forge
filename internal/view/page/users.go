package page

import (
	"fmt"

	"starter/internal/model"
	"starter/internal/view/layout"
	"starter/internal/view/partial/userpartial"
	"starter/pkg/components/patterns/flash"
	componenthtmx "starter/pkg/components/patterns/htmx"
	"starter/pkg/components/patterns/pager"
	"starter/pkg/components/ui/core"
	uidata "starter/pkg/components/ui/data"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// UserListProps configures the user management list page.
type UserListProps struct {
	Users           []model.User
	Pager           pager.Pager
	CSRFToken       string
	IsAuthenticated bool
	UserName        string
	Flash           []flash.Message
}

// UserListPage renders the full user table inside the base layout.
func UserListPage(p UserListProps) g.Node {
	rows := make([]g.Node, len(p.Users))
	for i, u := range p.Users {
		rows[i] = userpartial.UserRow(userpartial.UserRowProps{
			User:      u,
			CSRFToken: p.CSRFToken,
		})
	}

	return layout.Page(layout.Props{
		Title:           "Users",
		CSRFToken:       p.CSRFToken,
		IsAuthenticated: p.IsAuthenticated,
		UserName:        p.UserName,
		Flash:           p.Flash,
		Children: []g.Node{
			h.H1(h.Class("text-2xl font-bold mb-4"), g.Text("Users")),
			uidata.Table.Root(
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
			),
			pagination(p.Pager),
		},
	})
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
				Href:    fmt.Sprintf("/users?page=%d", p.PrevPage()),
				Variant: core.VariantGhost,
				Size:    core.SizeSm,
				Extra: componenthtmx.PaginatedTableLink(componenthtmx.PaginatedTableProps{
					Path:    "/users",
					Page:    p.PrevPage(),
					Target:  "#users-table",
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
				Href:    fmt.Sprintf("/users?page=%d", p.NextPage()),
				Variant: core.VariantGhost,
				Size:    core.SizeSm,
				Extra: componenthtmx.PaginatedTableLink(componenthtmx.PaginatedTableProps{
					Path:    "/users",
					Page:    p.NextPage(),
					Target:  "#users-table",
					Swap:    componenthtmx.SwapOuterHTML,
					PushURL: true,
				}),
			}),
		),
	)
}
