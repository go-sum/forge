// Example: safe to delete as a unit.
package userpartial

import (
	htmx "github.com/go-sum/componentry/patterns/htmx"
	"github.com/go-sum/componentry/ui/core"
	"github.com/go-sum/componentry/ui/data"
	authmodel "github.com/go-sum/auth/model"
	"github.com/go-sum/forge/internal/view"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// UserRowProps configures a read-only user table row.
type UserRowProps struct {
	User authmodel.User
}

// UserRow renders a <tr> with display data and HTMX-powered Edit/Delete actions.
// CSRF for HTMX mutations is provided by the body-level hx-headers attribute
// set in the page layout — no per-row token is needed.
func UserRow(req view.Request, p UserRowProps) g.Node {
	u := p.User
	id := u.ID.String()

	return data.Table.Row(data.RowProps{},
		h.ID("user-"+id),
		data.Table.Cell(g.Text(u.DisplayName)),
		data.Table.Cell(g.Text(u.Email)),
		data.Table.Cell(core.Badge(core.BadgeProps{
			Variant:  roleVariant(u.Role),
			Children: []g.Node{g.Text(u.Role)},
		})),
		data.Table.Cell(g.Text(u.CreatedAt.Format("2006-01-02"))),
		data.Table.Cell(
			h.Div(h.Class("flex justify-end gap-2"),
				core.Button(core.ButtonProps{
					Label:   "Edit",
					Variant: core.VariantGhost,
					Size:    core.SizeSm,
					Extra: htmx.Attrs(htmx.AttrsProps{
						Get:       req.Path("admin.user.edit", id),
						Target:    "closest tr",
						Swap:      htmx.SwapOuterHTML,
						Indicator: "#users-loading",
					}),
				}),
				core.Button(core.ButtonProps{
					Label:   "Delete",
					Variant: core.VariantDestructiveGhost,
					Size:    core.SizeSm,
					Type:    "button",
					Extra: htmx.Attrs(htmx.AttrsProps{
						Confirm:   "Delete " + u.DisplayName + "?",
						Delete:    req.Path("admin.user.delete", id),
						Target:    "closest tr",
						Swap:      "outerHTML swap:500ms",
						Indicator: "#users-loading",
					}),
				}),
			),
		),
	)
}

func roleVariant(role string) core.BadgeVariant {
	if role == authmodel.RoleAdmin {
		return core.BadgeDefault
	}
	return core.BadgeSecondary
}
