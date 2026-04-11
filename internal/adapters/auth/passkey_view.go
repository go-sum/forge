package authadapter

import (
	auth "github.com/go-sum/auth"
	htmx "github.com/go-sum/componentry/patterns/htmx"
	"github.com/go-sum/componentry/ui/core"
	uidata "github.com/go-sum/componentry/ui/data"
	uiform "github.com/go-sum/componentry/form"
	"github.com/go-sum/forge/internal/view"
	"github.com/go-sum/server/route"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

var _ auth.PasskeyPageRenderer = (*Renderer)(nil)

// PasskeyListPage renders the full passkey management page.
func (r *Renderer) PasskeyListPage(req auth.Request, data auth.PasskeyListData) g.Node {
	vreq := hostRequest(req)
	return vreq.Page(
		"Passkeys",
		h.Div(
			h.Class("space-y-6"),
			passkeyPageCard(vreq, data.CSRFToken),
			r.PasskeyListRegion(req, data),
		),
	)
}

// PasskeyListRegion renders the HTMX-replaceable passkey list region.
func (r *Renderer) PasskeyListRegion(req auth.Request, data auth.PasskeyListData) g.Node {
	vreq := hostRequest(req)
	return h.Div(
		h.ID("passkeys-list-region"),
		h.Class("space-y-4"),
		passkeyListContent(vreq, data),
	)
}

// PasskeyRow renders a single passkey row in the management table.
func (r *Renderer) PasskeyRow(req auth.Request, data auth.PasskeyRowData) g.Node {
	vreq := hostRequest(req)
	return renderPasskeyRow(vreq, data)
}

// PasskeyEditForm renders an inline edit form for renaming a passkey.
func (r *Renderer) PasskeyEditForm(req auth.Request, data auth.PasskeyRowData) g.Node {
	vreq := hostRequest(req)
	return renderPasskeyEditForm(vreq, data)
}

func passkeyPageCard(vreq view.Request, csrfToken string) g.Node {
	beginURL, _ := route.SafeReverse(vreq.Routes, "passkey.register.begin")
	finishURL, _ := route.SafeReverse(vreq.Routes, "passkey.register.finish")
	listURL, _ := route.SafeReverse(vreq.Routes, "passkey.list")

	return uidata.Card.Root(
		uidata.Card.Header(
			uidata.Card.Title(g.Text("Passkeys")),
			uidata.Card.Description(g.Text("Manage your registered passkeys for passwordless sign-in.")),
		),
		uidata.Card.Content(
			h.Div(
				h.Class("space-y-3"),
				h.P(
					g.Attr("data-passkey-error", ""),
					h.Class("hidden text-sm text-destructive"),
				),
				h.Div(
					h.Class("flex justify-end"),
					core.Button(core.ButtonProps{
						Label:   "Add passkey",
						Variant: core.VariantDefault,
						Type:    "button",
						Extra: []g.Node{
							g.Attr("data-passkey-register", ""),
							g.Attr("data-passkey-visible", ""),
							g.Attr("data-begin-url", beginURL),
							g.Attr("data-finish-url", finishURL),
							g.Attr("data-list-url", listURL),
							h.Class("hidden"),
						},
					}),
				),
			),
		),
	)
}

func passkeyListContent(req view.Request, data auth.PasskeyListData) g.Node {
	if len(data.Passkeys) == 0 {
		return uidata.Card.Root(
			uidata.Card.Content(
				h.Div(
					h.Class("flex flex-col items-center justify-center gap-2 py-10 text-center"),
					h.H2(h.Class("text-lg font-semibold"), g.Text("No passkeys registered.")),
					h.P(h.Class("text-sm text-muted-foreground"), g.Text("Add a passkey to sign in without a verification code.")),
				),
			),
		)
	}

	rows := make([]g.Node, len(data.Passkeys))
	for i, cred := range data.Passkeys {
		rows[i] = renderPasskeyRow(req, auth.PasskeyRowData{
			Passkey:   cred,
			CSRFToken: data.CSRFToken,
		})
	}

	return uidata.Table.Root(
		uidata.Table.Header(
			uidata.Table.Row(uidata.RowProps{},
				uidata.Table.Head(g.Text("Name")),
				uidata.Table.Head(g.Text("Type")),
				uidata.Table.Head(g.Text("Created")),
				uidata.Table.Head(g.Text("Last Used")),
				uidata.Table.Head(g.Text("Actions")),
			),
		),
		uidata.Table.Body(uidata.BodyProps{},
			g.Group(rows),
		),
	)
}

func renderPasskeyRow(req view.Request, data auth.PasskeyRowData) g.Node {
	cred := data.Passkey
	rowID := "passkey-row-" + cred.ID.String()
	deleteURL := req.Path("passkey.delete", cred.ID)
	renameURL := req.Path("passkey.rename.form", cred.ID)
	target := "#" + rowID

	lastUsed := "Never"
	if cred.LastUsedAt != nil {
		lastUsed = cred.LastUsedAt.Format("2006-01-02 15:04")
	}

	return uidata.Table.Row(uidata.RowProps{Extra: []g.Node{h.ID(rowID)}},
		uidata.Table.Cell(g.Text(cred.Name)),
		uidata.Table.Cell(g.Text(attachmentLabel(cred.Attachment))),
		uidata.Table.Cell(g.Text(cred.CreatedAt.Format("2006-01-02"))),
		uidata.Table.Cell(g.Text(lastUsed)),
		uidata.Table.Cell(
			h.Div(
				h.Class("flex items-center gap-2"),
				core.Button(core.ButtonProps{
					Label:   "Rename",
					Variant: core.VariantSecondary,
					Size:    core.SizeSm,
					Type:    "button",
					Extra: htmx.Attrs(htmx.AttrsProps{
						Get:    renameURL + "?edit=1",
						Target: target,
						Swap:   htmx.SwapOuterHTML,
					}),
				}),
				core.Button(core.ButtonProps{
					Label:   "Delete",
					Variant: core.VariantDestructiveGhost,
					Size:    core.SizeSm,
					Type:    "button",
					Extra: htmx.Attrs(htmx.AttrsProps{
						Delete:  deleteURL,
						Target:  target,
						Swap:    "delete",
						Confirm: "Remove this passkey?",
						Headers: map[string]string{"X-CSRF-Token": data.CSRFToken},
					}),
				}),
			),
		),
	)
}

func renderPasskeyEditForm(req view.Request, data auth.PasskeyRowData) g.Node {
	cred := data.Passkey
	rowID := "passkey-row-" + cred.ID.String()
	renameURL := req.Path("passkey.rename", cred.ID)
	rowURL := req.Path("passkey.row", cred.ID)
	target := "#" + rowID

	return h.Tr(
		h.ID(rowID),
		h.Class("border-b transition-colors hover:bg-muted/50"),
		h.Td(
			g.Attr("colspan", "5"),
			h.Class("p-2"),
			h.Form(
				g.Attr("hx-post", renameURL),
				g.Attr("hx-target", target),
				g.Attr("hx-swap", htmx.SwapOuterHTML),
				g.Attr("hx-headers", `{"X-CSRF-Token":"`+data.CSRFToken+`"}`),
				h.Class("flex items-center gap-2"),
				uiform.Input(uiform.InputProps{
					ID:    "passkey-name-" + cred.ID.String(),
					Name:  "name",
					Value: cred.Name,
					Extra: []g.Node{
						h.Required(),
						h.Placeholder("Passkey name"),
					},
				}),
				core.Button(core.ButtonProps{
					Label:   "Save",
					Type:    "submit",
					Variant: core.VariantDefault,
					Size:    core.SizeSm,
				}),
				core.Button(core.ButtonProps{
					Label:   "Cancel",
					Type:    "button",
					Variant: core.VariantSecondary,
					Size:    core.SizeSm,
					Extra: htmx.Attrs(htmx.AttrsProps{
						Get:    rowURL,
						Target: target,
						Swap:   htmx.SwapOuterHTML,
					}),
				}),
			),
		),
	)
}

func attachmentLabel(attachment string) string {
	switch attachment {
	case "platform":
		return "This device"
	case "cross-platform":
		return "Security key"
	default:
		return "Passkey"
	}
}
