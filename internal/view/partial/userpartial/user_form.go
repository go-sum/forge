// Package userpartial provides HTMX partial components for the user management table.
package userpartial

import (
	"starter/internal/model"
	"starter/internal/routes"
	"starter/internal/view"
	uiform "starter/pkg/components/form"
	componenthtmx "starter/pkg/components/patterns/htmx"
	"starter/pkg/components/ui/core"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// UserFormData configures the inline edit form row.
// Errors keys match Go struct field names (e.g. "Email", "DisplayName", "Role")
// as returned by pkgform.Submission.GetErrors() — no remapping needed at the call site.
type UserFormData struct {
	User   model.User
	Values model.UpdateUserInput
	Errors map[string][]string
}

// UserEditForm renders a <tr> containing a form for inline user editing.
// HTMX swaps this into the table via hx-get on the Edit button.
func UserEditForm(req view.Request, data UserFormData) g.Node {
	id := data.User.ID.String()
	emailID := "edit-email-" + id
	nameID := "edit-name-" + id
	roleID := "edit-role-" + id
	emailValue := data.User.Email
	if data.Values.Email != "" {
		emailValue = data.Values.Email
	}
	displayNameValue := data.User.DisplayName
	if data.Values.DisplayName != "" {
		displayNameValue = data.Values.DisplayName
	}
	roleValue := data.User.Role
	if data.Values.Role != "" {
		roleValue = data.Values.Role
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
				h.Input(h.Type("hidden"), h.Name("_csrf"), h.Value(req.CSRFToken)),
				view.FormError(data.Errors["_"]),
				uiform.Field(uiform.FieldProps{
					ID:     emailID,
					Label:  "Email",
					Errors: data.Errors["Email"],
					Control: uiform.Input(uiform.InputProps{
						ID:       emailID,
						Name:     "email",
						Type:     uiform.TypeEmail,
						Value:    emailValue,
						HasError: len(data.Errors["Email"]) > 0,
						Extra:    uiform.FieldControlAttrs(emailID, "", "", data.Errors["Email"]),
					}),
				}),
				uiform.Field(uiform.FieldProps{
					ID:     nameID,
					Label:  "Display Name",
					Errors: data.Errors["DisplayName"],
					Control: uiform.Input(uiform.InputProps{
						ID:       nameID,
						Name:     "display_name",
						Value:    displayNameValue,
						HasError: len(data.Errors["DisplayName"]) > 0,
						Extra:    uiform.FieldControlAttrs(nameID, "", "", data.Errors["DisplayName"]),
					}),
				}),
				uiform.Field(uiform.FieldProps{
					ID:     roleID,
					Label:  "Role",
					Errors: data.Errors["Role"],
					Control: uiform.Select(uiform.SelectProps{
						ID:       roleID,
						Name:     "role",
						Selected: roleValue,
						Options: []uiform.Option{
							{Value: model.RoleUser, Label: "User"},
							{Value: model.RoleAdmin, Label: "Admin"},
						},
						HasError: len(data.Errors["Role"]) > 0,
						Extra:    uiform.FieldControlAttrs(roleID, "", "", data.Errors["Role"]),
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

