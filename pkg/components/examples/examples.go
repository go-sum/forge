// Package examples renders a living reference of every visual component in
// pkg/components/. It produces a pure g.Node with no HTTP or internal/ imports,
// keeping pkg/ as a leaf node in the dependency graph.
package examples

import (
	uiform "starter/pkg/components/form"
	componenticons "starter/pkg/components/icons"
	iconrender "starter/pkg/components/icons/render"
	"starter/pkg/components/interactive/accordion"
	"starter/pkg/components/interactive/breadcrumb"
	"starter/pkg/components/interactive/dialog"
	"starter/pkg/components/interactive/dropdown"
	"starter/pkg/components/interactive/pagination"
	"starter/pkg/components/interactive/tabs"
	"starter/pkg/components/interactive/tooltip"
	"starter/pkg/components/ui/core"
	"starter/pkg/components/ui/data"
	"starter/pkg/components/ui/feedback"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// Page returns the full component showcase as a single renderable node.
func Page() g.Node {
	return h.Div(
		h.ID("top"),
		h.Class("max-w-4xl mx-auto space-y-12 py-8"),
		// Header
		h.Div(
			h.H1(h.Class("text-3xl font-bold mb-1"), g.Text("Component Examples")),
			h.P(h.Class("text-muted-foreground"), g.Text("Live reference for every visual component in pkg/components/.")),
		),
		// Table of Contents
		data.Card.Root(
			data.Card.Header(data.Card.Title(g.Text("Contents"))),
			data.Card.Content(
				h.Ul(h.Class("columns-3 gap-x-6 text-sm space-y-1"),
					tocItem("accordion", "Accordion"),
					tocItem("alerts", "Alerts"),
					tocItem("avatars", "Avatars"),
					tocItem("badges", "Badges"),
					tocItem("breadcrumb", "Breadcrumb"),
					tocItem("buttons", "Buttons"),
					tocItem("cards", "Cards"),
					tocItem("dialog", "Dialog"),
					tocItem("dropdown", "Dropdown"),
					tocItem("form-fields", "Form Fields"),
					tocItem("labels", "Labels"),
					tocItem("pagination", "Pagination"),
					tocItem("progress", "Progress"),
					tocItem("separators", "Separators"),
					tocItem("skeleton", "Skeleton"),
					tocItem("tables", "Tables"),
					tocItem("tabs", "Tabs"),
					tocItem("toast", "Toast"),
					tocItem("tooltip", "Tooltip"),
				),
			),
		),

		// ── Accordion ───────────────────────────────────
		section("accordion", "Accordion",
			example("Three items", accordion.Root(
				accordion.Item(
					accordion.Trigger(g.Text("Is it accessible?")),
					accordion.Content(g.Text("Yes. It uses native <details>/<summary> elements with WAI-ARIA semantics.")),
				),
				accordion.Item(
					accordion.Trigger(g.Text("Is it styled?")),
					accordion.Content(g.Text("Yes. It uses Tailwind utility classes with shadcn/ui design tokens.")),
				),
				accordion.Item(
					accordion.Trigger(g.Text("Is it animated?")),
					accordion.Content(g.Text("The chevron rotates on expand via CSS details[open] .details-chevron rule.")),
				),
			)),
		),

		// ── Alerts ──────────────────────────────────────
		section("alerts", "Alerts",
			h.Div(h.Class("grid grid-cols-2 gap-4"),
				example("Default (dismissible)", feedback.Alert.Root(
					feedback.AlertProps{Variant: feedback.AlertDefault, Dismissible: true},
					feedback.Alert.Title(g.Text("Note")),
					feedback.Alert.Description(g.Text("Here is some helpful information.")),
				)),
				example("Destructive (dismissible)", feedback.Alert.Root(
					feedback.AlertProps{Variant: feedback.AlertDestructive, Dismissible: true},
					feedback.Alert.Title(g.Text("Error")),
					feedback.Alert.Description(g.Text("Something went wrong. Please try again.")),
				)),
			),
		),

		// ── Avatars ──────────────────────────────────────
		section("avatars", "Avatars",
			h.Div(h.Class("grid grid-cols-2 gap-4"),
				example("With image + fallback", h.Div(
					h.Class("flex gap-4"),
					core.Avatar.Root(
						core.Avatar.Image("/public/img/svg/avatar.svg", "shadcn"),
						core.Avatar.Fallback(g.Text("CN")),
					),
					core.Avatar.Root(core.Avatar.Fallback(g.Text("AB"))),
				)),
				example("Lucide icon", core.Icon(iconrender.Props("lucide-icons", "circle-user", core.IconProps{
					Size:  "size-10",
					Label: "User account",
				}))),
			),
		),

		// ── Badges ──────────────────────────────────────
		section("badges", "Badges",
			example("Variants", h.Div(
				h.Class("flex flex-wrap gap-2"),
				feedback.Badge(feedback.BadgeProps{Children: []g.Node{g.Text("Default")}}),
				feedback.Badge(feedback.BadgeProps{Variant: feedback.BadgeSecondary, Children: []g.Node{g.Text("Secondary")}}),
				feedback.Badge(feedback.BadgeProps{Variant: feedback.BadgeDestructive, Children: []g.Node{g.Text("Destructive")}}),
				feedback.Badge(feedback.BadgeProps{Variant: feedback.BadgeOutline, Children: []g.Node{g.Text("Outline")}}),
			)),
			example("With icon", h.Div(
				h.Class("flex flex-wrap gap-2"),
				feedback.Badge(feedback.BadgeProps{Children: []g.Node{
					core.Icon(iconrender.Props("lucide-icons", "check", core.IconProps{})),
					g.Text("Verified"),
				}}),
				feedback.Badge(feedback.BadgeProps{Variant: feedback.BadgeDestructive, Children: []g.Node{
					core.Icon(iconrender.Props("lucide-icons", "x", core.IconProps{})),
					g.Text("Failed"),
				}}),
				feedback.Badge(feedback.BadgeProps{Variant: feedback.BadgeOutline, Children: []g.Node{
					core.Icon(iconrender.Props("lucide-icons", "clock", core.IconProps{})),
					g.Text("Pending"),
				}}),
			)),
		),

		// ── Breadcrumb ──────────────────────────────────
		section("breadcrumb", "Breadcrumb",
			example("Three-level path", breadcrumb.Root(
				breadcrumb.List(
					breadcrumb.Item(breadcrumb.Link("/", g.Text("Home"))),
					breadcrumb.Item(breadcrumb.Separator()),
					breadcrumb.Item(breadcrumb.Link("/users", g.Text("Users"))),
					breadcrumb.Item(breadcrumb.Separator()),
					breadcrumb.Item(breadcrumb.Page(g.Text("Alice Johnson"))),
				),
			)),
		),

		// ── Buttons ──────────────────────────────────────
		section("buttons", "Buttons",
			h.Div(h.Class("grid grid-cols-2 gap-4"),
				example("Variants", h.Div(
					h.Class("flex flex-wrap gap-2"),
					core.Button(core.Props{Label: "Default"}),
					core.Button(core.Props{Label: "Destructive", Variant: core.VariantDestructive}),
					core.Button(core.Props{Label: "Outline", Variant: core.VariantOutline}),
					core.Button(core.Props{Label: "Secondary", Variant: core.VariantSecondary}),
					core.Button(core.Props{Label: "Ghost", Variant: core.VariantGhost}),
					core.Button(core.Props{Label: "Link", Variant: core.VariantLink}),
				)),
				example("Sizes", h.Div(
					h.Class("flex flex-wrap items-center gap-2"),
					core.Button(core.Props{Label: "Large", Size: core.SizeLg}),
					core.Button(core.Props{Label: "Default"}),
					core.Button(core.Props{Label: "Small", Size: core.SizeSm}),
				)),
				example("Link (as <a>)", h.Div(
					h.Class("flex gap-2"),
					core.Button(core.Props{Label: "Go Home", Href: "/", Variant: core.VariantSecondary}),
					core.Button(core.Props{Label: "Users", Href: "/users", Variant: core.VariantGhost, Size: core.SizeSm}),
					core.Button(core.Props{Href: "/users", Variant: core.VariantOutline, Size: core.SizeSm, Children: []g.Node{
						core.Icon(iconrender.Props("lucide-icons", "users", core.IconProps{})),
						g.Text("Team"),
					}}),
				)),
				example("Disabled", h.Div(
					h.Class("flex gap-2"),
					core.Button(core.Props{Label: "Disabled", Disabled: true}),
					core.Button(core.Props{Label: "Disabled Outline", Variant: core.VariantOutline, Disabled: true}),
				)),
			),
		),

		// ── Cards ───────────────────────────────────────
		section("cards", "Cards",
			example("Full card anatomy", data.Card.Root(
				data.Card.Header(
					data.Card.Title(g.Text("Card Title")),
					data.Card.Description(g.Text("Optional description text goes here.")),
				),
				data.Card.Content(
					h.P(g.Text("This is the main body of the card. Cards compose header, content, and footer sub-components.")),
				),
				data.Card.Footer(
					core.Button(core.Props{Label: "Action", Size: core.SizeSm}),
				),
			)),
		),

		// ── Dialog ──────────────────────────────────────
		section("dialog", "Dialog",
			example("Modal dialog with trigger", dialog.Root(
				dialog.Trigger("example-dialog",
					core.Button(core.Props{Label: "Open Dialog"}),
				),
				dialog.Content("example-dialog",
					dialog.Header(
						dialog.Title(g.Text("Confirm Action")),
						dialog.Description(g.Text("This action cannot be undone. Are you sure you want to proceed?")),
					),
					dialog.Footer(
						dialog.Close(
							core.Button(core.Props{Label: "Cancel", Variant: core.VariantOutline}),
						),
						core.Button(core.Props{Label: "Confirm", Variant: core.VariantDestructive}),
					),
				),
			)),
		),

		// ── Dropdown ────────────────────────────────────
		section("dropdown", "Dropdown",
			example("Ghost button trigger", dropdown.Root(
				dropdown.Props{},
				dropdown.Trigger(
					core.Button(core.Props{Label: "Options ▾", Variant: core.VariantGhost}),
				),
				dropdown.Content(
					dropdown.Label("Account"),
					dropdown.Item("View Profile", "#", false),
					dropdown.Item("Edit Settings", "#", false),
					dropdown.Separator(),
					dropdown.Item("Sign Out", "#", false),
				),
			)),
		),

		// ── Form Fields ─────────────────────────────────
		section("form-fields", "Form Fields",
			example("Text input", uiform.Input(uiform.InputProps{
				ID:          "ex-text",
				Name:        "username",
				Placeholder: "e.g. alice",
			})),
			example("Email input (required)", uiform.Input(uiform.InputProps{
				ID:       "ex-email",
				Name:     "email",
				Type:     uiform.TypeEmail,
				Required: true,
			})),
			example("Input with error state", uiform.Input(uiform.InputProps{
				ID:       "ex-error",
				Name:     "password",
				Type:     uiform.TypePassword,
				Value:    "short",
				HasError: true,
			})),
			example("Select", uiform.Select(uiform.SelectProps{
				ID:       "ex-role",
				Name:     "role",
				Selected: "editor",
				Options: []uiform.Option{
					{Value: "admin", Label: "Admin"},
					{Value: "editor", Label: "Editor"},
					{Value: "viewer", Label: "Viewer"},
				},
			})),
			example("Checkbox (checked)", h.Label(
				h.Class("flex items-center gap-2 text-sm cursor-pointer"),
				uiform.Checkbox(uiform.CheckboxProps{
					ID:      "ex-cb-checked",
					Name:    "notify",
					Checked: true,
				}),
				g.Text("Send email notifications"),
			)),
			example("Radio button", h.Div(
				h.Class("flex flex-col gap-2"),
				h.Label(
					h.Class("flex items-center gap-2 text-sm cursor-pointer"),
					uiform.Radio(uiform.RadioProps{ID: "ex-radio-a", Name: "choice", Value: "a", Checked: true}),
					g.Text("Option A"),
				),
				h.Label(
					h.Class("flex items-center gap-2 text-sm cursor-pointer"),
					uiform.Radio(uiform.RadioProps{ID: "ex-radio-b", Name: "choice", Value: "b"}),
					g.Text("Option B"),
				),
				h.Label(
					h.Class("flex items-center gap-2 text-sm cursor-pointer"),
					uiform.Radio(uiform.RadioProps{ID: "ex-radio-c", Name: "choice", Value: "c"}),
					g.Text("Option C"),
				),
			)),
			example("Switch (toggle)", h.Label(
				h.Class("flex items-center gap-2 text-sm cursor-pointer"),
				uiform.Switch(uiform.SwitchProps{
					ID:      "ex-switch",
					Name:    "enabled",
					Checked: true,
				}),
				g.Text("Enable feature"),
			)),
			example("Textarea", uiform.Textarea(uiform.TextareaProps{
				ID:          "ex-bio",
				Name:        "bio",
				Placeholder: "Tell us about yourself…",
				Rows:        4,
			})),
		),

		// ── Labels ──────────────────────────────────────
		section("labels", "Labels",
			example("Default", core.Label(core.LabelProps{For: "ex-input"}, g.Text("Email address"))),
			example("Error state", core.Label(core.LabelProps{For: "ex-input-err", Error: "Required"}, g.Text("Password"))),
		),

		// ── Pagination ──────────────────────────────────
		section("pagination", "Pagination",
			example("Five-page example (page 3 active)", pagination.Root(
				pagination.Content(
					pagination.Item(pagination.Previous("/users?page=2", false)),
					pagination.Item(pagination.Link("/users?page=1", false, g.Text("1"))),
					pagination.Item(pagination.Link("/users?page=2", false, g.Text("2"))),
					pagination.Item(pagination.Link("/users?page=3", true, g.Text("3"))),
					pagination.Item(pagination.Link("/users?page=4", false, g.Text("4"))),
					pagination.Item(pagination.Link("/users?page=5", false, g.Text("5"))),
					pagination.Item(pagination.Next("/users?page=4", false)),
				),
			)),
		),

		// ── Progress ────────────────────────────────────
		section("progress", "Progress",
			example("Default 60%", feedback.Progress(feedback.ProgressProps{Value: 60, Label: "Loading…", ShowValue: true})),
			example("Success 100%", feedback.Progress(feedback.ProgressProps{Variant: feedback.ProgressSuccess, Value: 100, ShowValue: true})),
			example("Danger 25%", feedback.Progress(feedback.ProgressProps{Variant: feedback.ProgressDanger, Value: 25, ShowValue: true})),
			example("Small", feedback.Progress(feedback.ProgressProps{Size: feedback.ProgressSm, Value: 40})),
		),

		// ── Separators ──────────────────────────────────
		section("separators", "Separators",
			example("Horizontal (plain)", core.Separator(core.SeparatorProps{})),
			example("Horizontal with label", core.Separator(core.SeparatorProps{Label: "OR"})),
			example("Dashed", core.Separator(core.SeparatorProps{Decoration: core.DecorationDashed})),
		),

		// ── Skeleton ────────────────────────────────────
		section("skeleton", "Skeleton",
			example("Loading placeholders", h.Div(
				h.Class("space-y-2"),
				core.Skeleton(h.Class("h-4 w-[250px]")),
				core.Skeleton(h.Class("h-4 w-[200px]")),
				core.Skeleton(h.Class("h-4 w-[150px]")),
			)),
		),

		// ── Tables ──────────────────────────────────────
		section("tables", "Tables",
			example("Table with header/body/actions", data.Table.Root(
				data.Table.Header(
					data.Table.Row(false,
						data.Table.Head(g.Text("Name")),
						data.Table.Head(g.Text("Role")),
						data.Table.Head(g.Text("Status")),
						data.Table.Head(g.Text("")),
					),
				),
				data.Table.Body(
					data.Table.Row(false,
						data.Table.Cell(g.Text("Alice Johnson")),
						data.Table.Cell(g.Text("Admin")),
						data.Table.Cell(feedback.Badge(feedback.BadgeProps{Children: []g.Node{g.Text("Active")}})),
						data.Table.Cell(
							h.Div(h.Class("flex justify-end gap-2"),
								core.Button(core.Props{Label: "Edit", Variant: core.VariantGhost, Size: core.SizeSm}),
								core.Button(core.Props{Label: "Delete", Variant: core.VariantDestructive, Size: core.SizeSm}),
							),
						),
					),
					data.Table.Row(false,
						data.Table.Cell(g.Text("Bob Smith")),
						data.Table.Cell(g.Text("Editor")),
						data.Table.Cell(feedback.Badge(feedback.BadgeProps{Variant: feedback.BadgeSecondary, Children: []g.Node{g.Text("Inactive")}})),
						data.Table.Cell(
							h.Div(h.Class("flex justify-end gap-2"),
								core.Button(core.Props{Label: "Edit", Variant: core.VariantGhost, Size: core.SizeSm}),
								core.Button(core.Props{Label: "Delete", Variant: core.VariantDestructive, Size: core.SizeSm}),
							),
						),
					),
				),
				data.Table.Caption(g.Text("A list of team members.")),
			)),
		),

		// ── Tabs ────────────────────────────────────────
		section("tabs", "Tabs",
			example("Three-tab panel", tabs.Root("account",
				tabs.List(
					tabs.Trigger("account", true, g.Text("Account")),
					tabs.Trigger("password", false, g.Text("Password")),
					tabs.Trigger("settings", false, g.Text("Settings")),
				),
				tabs.Content("account", true,
					data.Card.Root(
						data.Card.Header(data.Card.Title(g.Text("Account"))),
						data.Card.Content(h.P(g.Text("Manage your account settings here."))),
					),
				),
				tabs.Content("password", false,
					data.Card.Root(
						data.Card.Header(data.Card.Title(g.Text("Password"))),
						data.Card.Content(h.P(g.Text("Change your password here."))),
					),
				),
				tabs.Content("settings", false,
					data.Card.Root(
						data.Card.Header(data.Card.Title(g.Text("Settings"))),
						data.Card.Content(h.P(g.Text("Manage your preferences here."))),
					),
				),
			)),
		),

		// ── Toast ───────────────────────────────────────
		section("toast", "Toast",
			example("Variants", h.Div(
				h.Class("grid grid-cols-2 gap-3"),
				toastPreview("", "Event created", "Your event has been created."),
				toastPreview("success", "Success", "Changes saved successfully."),
				toastPreview("error", "Error", "Something went wrong."),
				toastPreview("warning", "Warning", "This action is irreversible."),
				toastPreview("info", "Info", "New updates are available."),
			)),
			example("Dismissible (click ×)", h.Div(
				h.Class("grid grid-cols-2 gap-3"),
				toastPreview("", "Notification", "Click the × button to dismiss."),
				toastPreview("success", "Saved", "Your changes have been saved."),
			)),
			example("Interactive — click to trigger (auto-dismisses after 5s)", h.Div(
				h.Class("flex flex-wrap gap-2"),
				toastTriggerButton("toast-tmpl-default", "Default"),
				toastTriggerButton("toast-tmpl-success", "Success"),
				toastTriggerButton("toast-tmpl-error", "Error"),
				toastTriggerButton("toast-tmpl-warning", "Warning"),
				toastTriggerButton("toast-tmpl-info", "Info"),
				toastTemplate("toast-tmpl-default", "", "Event created", "Your event has been created."),
				toastTemplate("toast-tmpl-success", "success", "Success", "Changes saved successfully."),
				toastTemplate("toast-tmpl-error", "error", "Error", "Something went wrong."),
				toastTemplate("toast-tmpl-warning", "warning", "Warning", "This action is irreversible."),
				toastTemplate("toast-tmpl-info", "info", "Info", "New updates are available."),
			)),
		),

		// ── Tooltip ─────────────────────────────────────
		section("tooltip", "Tooltip",
			example("Hover for tooltip", tooltip.Root(
				tooltip.Trigger(
					core.Button(core.Props{Label: "Hover me", Variant: core.VariantOutline}),
				),
				tooltip.Content(g.Text("This is a tooltip")),
			)),
		),
	)
}

// section renders an anchored <section> with a heading and divider.
func section(id, title string, content ...g.Node) g.Node {
	return h.Section(
		h.ID(id),
		h.Div(
			h.Class("flex items-center justify-between mb-4 scroll-mt-6"),
			h.H2(
				h.Class("text-xl font-semibold"),
				h.A(h.Href("#"+id), h.Class("hover:underline"), g.Text(title)),
			),
			h.A(
				h.Href("#top"),
				h.Class("text-xs text-muted-foreground hover:text-foreground hover:underline"),
				core.Icon(iconrender.PropsFor(componenticons.ChevronsUp, core.IconProps{Size: "size-4 shrink-0"})),
				// g.Text("↑ Back to top"),
			),
		),
		h.Div(h.Class("space-y-4"), g.Group(content)),
		h.Hr(h.Class("mt-8 border-border")),
	)
}

// example renders a named example box with a label and the component.
func example(name string, node g.Node) g.Node {
	return h.Div(
		h.Class("border border-border rounded-lg p-4"),
		h.P(h.Class("text-xs font-mono text-muted-foreground mb-3"), g.Text(name)),
		node,
	)
}

// tocItem renders a single table-of-contents anchor link.
func tocItem(id, label string) g.Node {
	return h.Li(
		h.Class("break-inside-avoid"),
		h.A(h.Href("#"+id), h.Class("text-muted-foreground hover:text-foreground hover:underline"), g.Text(label)),
	)
}

// toastVariantClass returns the colour classes for a toast variant.
func toastVariantClass(variant string) string {
	switch variant {
	case "success":
		return "border-success/20 bg-success/10 text-success"
	case "error":
		return "border-destructive/20 bg-destructive/10 text-destructive"
	case "warning":
		return "border-warning/20 bg-warning/10 text-warning"
	case "info":
		return "border-blue-200 bg-blue-50 text-blue-900"
	default:
		return "border-border bg-background text-foreground"
	}
}

// toastPreview renders an inline (non-fixed) dismissible toast for variant showcase.
func toastPreview(variant, title, desc string) g.Node {
	return h.Div(
		h.Class("relative rounded-lg border p-4 shadow-md "+toastVariantClass(variant)),
		g.Attr("data-dismissible", ""),
		h.P(h.Class("font-medium text-sm"), g.Text(title)),
		h.P(h.Class("text-sm mt-1 opacity-80"), g.Text(desc)),
		h.Button(
			g.Attr("data-dismiss", ""),
			h.Class("absolute top-2 right-2 opacity-50 hover:opacity-100 transition-opacity text-xs"),
			h.Type("button"),
			g.Text("×"),
		),
	)
}

// toastTriggerButton renders a button that clones a <template> toast into #toast-container.
func toastTriggerButton(templateID, label string) g.Node {
	return core.Button(core.Props{
		Label:   label,
		Variant: core.VariantOutline,
		Size:    core.SizeSm,
		Extra:   []g.Node{g.Attr("data-toast-trigger", templateID)},
	})
}

// toastTemplate renders a hidden <template> containing a toast for JS cloning.
func toastTemplate(id, variant, title, desc string) g.Node {
	return g.El("template", h.ID(id),
		h.Div(
			h.Class("relative rounded-lg border p-4 shadow-md "+toastVariantClass(variant)),
			g.Attr("data-dismissible", ""),
			h.P(h.Class("font-medium text-sm"), g.Text(title)),
			h.P(h.Class("text-sm mt-1 opacity-80"), g.Text(desc)),
			h.Button(
				g.Attr("data-dismiss", ""),
				h.Class("absolute top-2 right-2 opacity-50 hover:opacity-100 transition-opacity text-xs"),
				h.Type("button"),
				g.Text("×"),
			),
		),
	)
}
