// Package userpartial provides HTMX partial components for the user management table.
package userpartial

import (
	"fmt"

	"starter/internal/model"
	uiform "starter/pkg/components/form"
	componenthtmx "starter/pkg/components/patterns/htmx"
	"starter/pkg/components/ui/core"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// UserFormProps configures the inline edit form row.
// Errors is map[string][]string (not *form.Submission) to keep this partial
// free of pkg/form imports — callers extract errors before passing them in.
type UserFormProps struct {
	User      model.User
	CSRFToken string
	Errors    map[string][]string
}

// UserEditForm renders a <tr> containing a form for inline user editing.
// HTMX swaps this into the table via hx-get on the Edit button.
func UserEditForm(p UserFormProps) g.Node {
	id := p.User.ID.String()
	emailID := "edit-email-" + id
	nameID := "edit-name-" + id
	roleID := "edit-role-" + id

	return h.Tr(
		h.ID("user-"+id),
		h.Td(
			h.ColSpan("5"),
			h.Form(
				g.Group(componenthtmx.Attrs(componenthtmx.AttrsProps{
					Put:    fmt.Sprintf("/users/%s", id),
					Target: "closest tr",
					Swap:   componenthtmx.SwapOuterHTML,
				})),
				h.Class("flex flex-wrap gap-3 items-end p-2"),
				h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(p.CSRFToken)),
				uiform.Field(uiform.FieldProps{
					ID:     emailID,
					Label:  "Email",
					Errors: p.Errors["email"],
					Control: uiform.Input(uiform.InputProps{
						ID:       emailID,
						Name:     "email",
						Type:     uiform.TypeEmail,
						Value:    p.User.Email,
						HasError: len(p.Errors["email"]) > 0,
						Extra:    uiform.FieldControlAttrs(emailID, "", "", p.Errors["email"]),
					}),
				}),
				uiform.Field(uiform.FieldProps{
					ID:     nameID,
					Label:  "Display Name",
					Errors: p.Errors["display_name"],
					Control: uiform.Input(uiform.InputProps{
						ID:       nameID,
						Name:     "display_name",
						Value:    p.User.DisplayName,
						HasError: len(p.Errors["display_name"]) > 0,
						Extra:    uiform.FieldControlAttrs(nameID, "", "", p.Errors["display_name"]),
					}),
				}),
				uiform.Field(uiform.FieldProps{
					ID:     roleID,
					Label:  "Role",
					Errors: p.Errors["role"],
					Control: uiform.Select(uiform.SelectProps{
						ID:       roleID,
						Name:     "role",
						Selected: p.User.Role,
						Options: []uiform.Option{
							{Value: "user", Label: "User"},
							{Value: "admin", Label: "Admin"},
						},
						HasError: len(p.Errors["role"]) > 0,
						Extra:    uiform.FieldControlAttrs(roleID, "", "", p.Errors["role"]),
					}),
				}),
				h.Div(
					h.Class("flex gap-2 mt-4"),
					core.Button(core.ButtonProps{
						Label: "Save",
						Size:  core.SizeSm,
						Type:  "submit",
					}),
					core.Button(core.ButtonProps{
						Label:   "Cancel",
						Variant: core.VariantGhost,
						Size:    core.SizeSm,
						Extra: componenthtmx.Attrs(componenthtmx.AttrsProps{
							Get:    fmt.Sprintf("/users/%s/row", id),
							Target: "closest tr",
							Swap:   componenthtmx.SwapOuterHTML,
						}),
					}),
				),
			),
		),
	)
}
