// Example: safe to delete as a unit.
// Package userpartial provides HTMX partial components for the user management table.
package userpartial

import (
	uiform "github.com/go-sum/componentry/form"
	htmx "github.com/go-sum/componentry/patterns/htmx"
	"github.com/go-sum/componentry/ui/core"
	authmodel "github.com/go-sum/auth/model"
	"github.com/go-sum/forge/internal/view"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// UserFormData configures the inline edit form row.
// Errors keys match Go struct field names (e.g. "Email", "DisplayName", "Role")
// as returned by patterns/form.Submission.GetErrors() — no remapping needed at the call site.
type UserFormData struct {
	User   authmodel.User
	Values authmodel.UpdateUserInput
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
				g.Group(htmx.Attrs(htmx.AttrsProps{
					Put:       req.Path("admin.user.update", id),
					Target:    "closest tr",
					Swap:      htmx.SwapOuterHTML,
					Indicator: "#users-loading",
				})),
				h.Class("grid gap-4 p-3 sm:grid-cols-2 xl:grid-cols-[minmax(0,1.4fr)_minmax(0,1.1fr)_12rem_auto] xl:items-end"),
				h.Input(h.Type("hidden"), h.Name(req.CSRFFieldName), h.Value(req.CSRFToken)),
				h.Div(
					h.Class("sm:col-span-2 xl:col-span-4"),
					view.FormError(data.Errors["_"]),
				),
				uiform.Field(uiform.FieldProps{
					ID:     emailID,
					Label:  "Email",
					Errors: data.Errors["Email"],
					Extra:  []g.Node{h.Class("min-w-0")},
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
					Extra:  []g.Node{h.Class("min-w-0")},
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
					Extra:  []g.Node{h.Class("min-w-0")},
					Control: uiform.Select(uiform.SelectProps{
						ID:       roleID,
						Name:     "role",
						Selected: roleValue,
						Options: []uiform.Option{
							{Value: authmodel.RoleUser, Label: "User"},
							{Value: authmodel.RoleAdmin, Label: "Admin"},
						},
						HasError: len(data.Errors["Role"]) > 0,
						Extra:    uiform.FieldControlAttrs(roleID, "", "", data.Errors["Role"]),
					}),
				}),
				h.Div(
					h.Class("flex flex-wrap gap-2 sm:col-span-2 xl:col-span-1 xl:justify-end"),
					core.Button(core.ButtonProps{
						Label: "Save",
						Size:  core.SizeSm,
						Type:  "submit",
					}),
					core.Button(core.ButtonProps{
						Label:   "Cancel",
						Variant: core.VariantGhost,
						Size:    core.SizeSm,
						Extra: htmx.Attrs(htmx.AttrsProps{
							Get:       req.Path("admin.user.row", id),
							Target:    "closest tr",
							Swap:      htmx.SwapOuterHTML,
							Indicator: "#users-loading",
						}),
					}),
				),
			),
		),
	)
}
