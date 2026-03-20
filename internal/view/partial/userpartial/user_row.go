package userpartial

import (
	"fmt"

	"starter/internal/model"
	componenthtmx "starter/pkg/components/patterns/htmx"
	"starter/pkg/components/ui/core"
	"starter/pkg/components/ui/data"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// UserRowProps configures a read-only user table row.
type UserRowProps struct {
	User      model.User
	CSRFToken string
}

// UserRow renders a <tr> with display data and HTMX-powered Edit/Delete actions.
func UserRow(p UserRowProps) g.Node {
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
					Extra: componenthtmx.Attrs(componenthtmx.AttrsProps{
						Get:    fmt.Sprintf("/users/%s/edit", id),
						Target: "closest tr",
						Swap:   componenthtmx.SwapOuterHTML,
					}),
				}),
				h.Form(
					h.Method("post"),
					h.Action(fmt.Sprintf("/users/%s", id)),
					h.Input(h.Type("hidden"), h.Name("_method"), h.Value("DELETE")),
					h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(p.CSRFToken)),
					core.Button(core.ButtonProps{
						Label:   "Delete",
						Variant: core.VariantDestructive,
						Size:    core.SizeSm,
						Type:    "submit",
						Extra: componenthtmx.Attrs(componenthtmx.AttrsProps{
							Confirm: fmt.Sprintf("Delete %s?", u.DisplayName),
							Delete:  fmt.Sprintf("/users/%s", id),
							Target:  "closest tr",
							Swap:    "outerHTML swap:500ms",
						}),
					}),
				),
			),
		),
	)
}

func roleVariant(role string) core.BadgeVariant {
	if role == "admin" {
		return core.BadgeDefault
	}
	return core.BadgeSecondary
}
