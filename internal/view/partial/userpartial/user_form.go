// Package userpartial provides HTMX partial components for the user management table.
package userpartial

import (
	"starter/internal/model"
	"starter/internal/routes"
	uiform "starter/pkg/components/form"
	componenthtmx "starter/pkg/components/patterns/htmx"
	"starter/pkg/components/ui/core"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// UserFormProps configures the inline edit form row.
// Errors keys match Go struct field names (e.g. "Email", "DisplayName", "Role")
// as returned by pkgform.Submission.GetErrors() — no remapping needed at the call site.
type UserFormProps struct {
	User      model.User
	Values    model.UpdateUserInput
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
	emailValue := p.User.Email
	if p.Values.Email != "" {
		emailValue = p.Values.Email
	}
	displayNameValue := p.User.DisplayName
	if p.Values.DisplayName != "" {
		displayNameValue = p.Values.DisplayName
	}
	roleValue := p.User.Role
	if p.Values.Role != "" {
		roleValue = p.Values.Role
	}

	return h.Tr(
		h.ID("user-"+id),
		h.Td(
			h.ColSpan("5"),
			h.Form(
				g.Group(componenthtmx.Attrs(componenthtmx.AttrsProps{
					Put:    routes.UserPath(id),
					Target: "closest tr",
					Swap:   componenthtmx.SwapOuterHTML,
				})),
				h.Class("flex flex-wrap gap-3 items-end p-2"),
				h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(p.CSRFToken)),
				formNotice(p.Errors["_"]),
				uiform.Field(uiform.FieldProps{
					ID:     emailID,
					Label:  "Email",
					Errors: p.Errors["Email"],
					Control: uiform.Input(uiform.InputProps{
						ID:       emailID,
						Name:     "email",
						Type:     uiform.TypeEmail,
						Value:    emailValue,
						HasError: len(p.Errors["Email"]) > 0,
						Extra:    uiform.FieldControlAttrs(emailID, "", "", p.Errors["Email"]),
					}),
				}),
				uiform.Field(uiform.FieldProps{
					ID:     nameID,
					Label:  "Display Name",
					Errors: p.Errors["DisplayName"],
					Control: uiform.Input(uiform.InputProps{
						ID:       nameID,
						Name:     "display_name",
						Value:    displayNameValue,
						HasError: len(p.Errors["DisplayName"]) > 0,
						Extra:    uiform.FieldControlAttrs(nameID, "", "", p.Errors["DisplayName"]),
					}),
				}),
				uiform.Field(uiform.FieldProps{
					ID:     roleID,
					Label:  "Role",
					Errors: p.Errors["Role"],
					Control: uiform.Select(uiform.SelectProps{
						ID:       roleID,
						Name:     "role",
						Selected: roleValue,
						Options: []uiform.Option{
							{Value: "user", Label: "User"},
							{Value: "admin", Label: "Admin"},
						},
						HasError: len(p.Errors["Role"]) > 0,
						Extra:    uiform.FieldControlAttrs(roleID, "", "", p.Errors["Role"]),
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
							Get:    routes.UserRowPath(id),
							Target: "closest tr",
							Swap:   componenthtmx.SwapOuterHTML,
						}),
					}),
				),
			),
		),
	)
}

func formNotice(messages []string) g.Node {
	if len(messages) == 0 {
		return g.Text("")
	}
	return h.Div(
		h.Class("w-full rounded-md border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm text-destructive"),
		g.Text(messages[0]),
	)
}
