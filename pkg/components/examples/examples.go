// Package examples renders a living reference of every visual component in
// pkg/components/. It produces a pure g.Node with no HTTP or internal/ imports,
// keeping pkg/ as a leaf node in the dependency graph.
package examples

import (
	componentassets "starter/pkg/components/assets"
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
					tocItem("popover", "Popover"),
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
				example("Image and placeholder", h.Div(
					h.Class("flex gap-4"),
					core.Avatar.Image(componentassets.PublicPath("img/svg/avatar.svg"), "shadcn"),
					core.Avatar.Fallback(g.Text("AB")),
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
				core.Badge(core.BadgeProps{Children: []g.Node{g.Text("Default")}}),
				core.Badge(core.BadgeProps{Variant: core.BadgeSecondary, Children: []g.Node{g.Text("Secondary")}}),
				core.Badge(core.BadgeProps{Variant: core.BadgeDestructive, Children: []g.Node{g.Text("Destructive")}}),
				core.Badge(core.BadgeProps{Variant: core.BadgeOutline, Children: []g.Node{g.Text("Outline")}}),
			)),
			example("With icon", h.Div(
				h.Class("flex flex-wrap gap-2"),
				core.Badge(core.BadgeProps{Children: []g.Node{
					core.Icon(iconrender.Props("lucide-icons", "check", core.IconProps{})),
					g.Text("Verified"),
				}}),
				core.Badge(core.BadgeProps{Variant: core.BadgeDestructive, Children: []g.Node{
					core.Icon(iconrender.Props("lucide-icons", "x", core.IconProps{})),
					g.Text("Failed"),
				}}),
				core.Badge(core.BadgeProps{Variant: core.BadgeOutline, Children: []g.Node{
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
					core.Button(core.ButtonProps{Label: "Default"}),
					core.Button(core.ButtonProps{Label: "Destructive", Variant: core.VariantDestructive}),
					core.Button(core.ButtonProps{Label: "Outline", Variant: core.VariantOutline}),
					core.Button(core.ButtonProps{Label: "Secondary", Variant: core.VariantSecondary}),
					core.Button(core.ButtonProps{Label: "Ghost", Variant: core.VariantGhost}),
					core.Button(core.ButtonProps{Label: "Link", Variant: core.VariantLink}),
				)),
				example("Sizes", h.Div(
					h.Class("flex flex-wrap items-center gap-2"),
					core.Button(core.ButtonProps{Label: "Large", Size: core.SizeLg}),
					core.Button(core.ButtonProps{Label: "Default"}),
					core.Button(core.ButtonProps{Label: "Small", Size: core.SizeSm}),
				)),
				example("Link (as <a>)", h.Div(
					h.Class("flex gap-2"),
					core.Button(core.ButtonProps{Label: "Go Home", Href: "/", Variant: core.VariantSecondary}),
					core.Button(core.ButtonProps{Label: "Users", Href: "/users", Variant: core.VariantGhost, Size: core.SizeSm}),
					core.Button(core.ButtonProps{Href: "/users", Variant: core.VariantOutline, Size: core.SizeSm, Children: []g.Node{
						core.Icon(iconrender.Props("lucide-icons", "users", core.IconProps{})),
						g.Text("Team"),
					}}),
				)),
				example("Disabled", h.Div(
					h.Class("flex gap-2"),
					core.Button(core.ButtonProps{Label: "Disabled", Disabled: true}),
					core.Button(core.ButtonProps{Label: "Disabled Outline", Variant: core.VariantOutline, Disabled: true}),
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
					core.Button(core.ButtonProps{Label: "Action", Size: core.SizeSm}),
				),
			)),
		),

		// ── Dialog ──────────────────────────────────────
		section("dialog", "Dialog",
			example("Modal dialog with trigger", dialog.Root(
				dialog.Trigger("example-dialog",
					core.Button(core.ButtonProps{Label: "Open Dialog"}),
				),
				dialog.Content("example-dialog",
					dialog.Header(
						dialog.Title("example-dialog", g.Text("Confirm Action")),
						dialog.Description("example-dialog", g.Text("This action cannot be undone. Are you sure you want to proceed?")),
					),
					dialog.Footer(
						dialog.Close(
							core.Button(core.ButtonProps{Label: "Cancel", Variant: core.VariantOutline}),
						),
						core.Button(core.ButtonProps{Label: "Confirm", Variant: core.VariantDestructive}),
					),
				),
			)),
		),

		// ── Dropdown ────────────────────────────────────
		section("dropdown", "Dropdown",
			example("Native summary trigger", dropdown.Root(
				dropdown.Props{},
				dropdown.Trigger(dropdown.TriggerProps{}, g.Text("Options")),
				dropdown.Content(
					dropdown.Label("Account"),
					dropdown.Item(dropdown.ItemProps{Label: "View Profile", Href: "#"}),
					dropdown.Item(dropdown.ItemProps{Label: "Edit Settings", Href: "#"}),
					dropdown.Separator(),
					dropdown.Item(dropdown.ItemProps{Label: "Sign Out", Href: "#"}),
				),
			)),
		),
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
			example("FieldSet — radio group", uiform.FieldSet(uiform.FieldSetProps{
				ID:     "ex-contact",
				Legend: "Preferred contact",
			},
				h.Label(h.Class("flex items-center gap-2 text-sm cursor-pointer"),
					uiform.Radio(uiform.RadioProps{ID: "ex-contact-email", Name: "contact", Value: "email", Checked: true}),
					g.Text("Email"),
				),
				h.Label(h.Class("flex items-center gap-2 text-sm cursor-pointer"),
					uiform.Radio(uiform.RadioProps{ID: "ex-contact-phone", Name: "contact", Value: "phone"}),
					g.Text("Phone"),
				),
				h.Label(h.Class("flex items-center gap-2 text-sm cursor-pointer"),
					uiform.Radio(uiform.RadioProps{ID: "ex-contact-post", Name: "contact", Value: "post"}),
					g.Text("Post"),
				),
			)),
			example("FieldSet — disabled group", uiform.FieldSet(uiform.FieldSetProps{
				ID:       "ex-contact-disabled",
				Legend:   "Preferred contact",
				Disabled: true,
			},
				h.Label(h.Class("flex items-center gap-2 text-sm cursor-pointer"),
					uiform.Radio(uiform.RadioProps{ID: "ex-cd-email", Name: "contact-disabled", Value: "email", Checked: true}),
					g.Text("Email"),
				),
				h.Label(h.Class("flex items-center gap-2 text-sm cursor-pointer"),
					uiform.Radio(uiform.RadioProps{ID: "ex-cd-phone", Name: "contact-disabled", Value: "phone"}),
					g.Text("Phone"),
				),
				h.Label(h.Class("flex items-center gap-2 text-sm cursor-pointer"),
					uiform.Radio(uiform.RadioProps{ID: "ex-cd-post", Name: "contact-disabled", Value: "post"}),
					g.Text("Post"),
				),
			)),
			example("Select with opt-groups", uiform.Select(uiform.SelectProps{
				ID:       "ex-role-grouped",
				Name:     "role",
				Selected: "admin",
				Groups: []uiform.OptGroup{
					{Label: "Admin roles", Options: []uiform.Option{
						{Value: "admin", Label: "Admin"},
						{Value: "superadmin", Label: "Super Admin"},
					}},
					{Label: "Member roles", Options: []uiform.Option{
						{Value: "editor", Label: "Editor"},
						{Value: "viewer", Label: "Viewer"},
					}},
				},
			})),
			example("File upload (single)", uiform.FileUpload(uiform.FileUploadProps{
				ID:     "ex-upload",
				Name:   "file",
				Accept: "image/*,application/pdf",
				Prompt: "Drop an image or PDF, or click to browse",
			})),
			example("File upload (multiple)", uiform.FileUpload(uiform.FileUploadProps{
				ID:       "ex-upload-multi",
				Name:     "files",
				Multiple: true,
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
					data.Table.Row(data.RowProps{},
						data.Table.Head(g.Text("Name")),
						data.Table.Head(g.Text("Role")),
						data.Table.Head(g.Text("Status")),
						data.Table.Head(g.Text("")),
					),
				),
				data.Table.Body(
					data.Table.Row(data.RowProps{},
						data.Table.Cell(g.Text("Alice Johnson")),
						data.Table.Cell(g.Text("Admin")),
						data.Table.Cell(core.Badge(core.BadgeProps{Children: []g.Node{g.Text("Active")}})),
						data.Table.Cell(
							h.Div(h.Class("flex justify-end gap-2"),
								core.Button(core.ButtonProps{Label: "Edit", Variant: core.VariantGhost, Size: core.SizeSm}),
								core.Button(core.ButtonProps{Label: "Delete", Variant: core.VariantDestructive, Size: core.SizeSm}),
							),
						),
					),
					data.Table.Row(data.RowProps{},
						data.Table.Cell(g.Text("Bob Smith")),
						data.Table.Cell(g.Text("Editor")),
						data.Table.Cell(core.Badge(core.BadgeProps{Variant: core.BadgeSecondary, Children: []g.Node{g.Text("Inactive")}})),
						data.Table.Cell(
							h.Div(h.Class("flex justify-end gap-2"),
								core.Button(core.ButtonProps{Label: "Edit", Variant: core.VariantGhost, Size: core.SizeSm}),
								core.Button(core.ButtonProps{Label: "Delete", Variant: core.VariantDestructive, Size: core.SizeSm}),
							),
						),
					),
				),
				data.Table.Caption(g.Text("A list of team members.")),
			)),
		),

		// ── Tabs ────────────────────────────────────────
		section("tabs", "Tabs",
			example("Three-tab panel", tabs.Root("account-tabs", "account",
				tabs.List(
					tabs.Trigger("account-tabs", "account", true, g.Text("Account")),
					tabs.Trigger("account-tabs", "password", false, g.Text("Password")),
					tabs.Trigger("account-tabs", "settings", false, g.Text("Settings")),
				),
				tabs.Content("account-tabs", "account", true,
					data.Card.Root(
						data.Card.Header(data.Card.Title(g.Text("Account"))),
						data.Card.Content(h.P(g.Text("Manage your account settings here."))),
					),
				),
				tabs.Content("account-tabs", "password", false,
					data.Card.Root(
						data.Card.Header(data.Card.Title(g.Text("Password"))),
						data.Card.Content(h.P(g.Text("Change your password here."))),
					),
				),
				tabs.Content("account-tabs", "settings", false,
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
				h.Class("flex flex-col gap-2"),
				feedback.Toast(feedback.ToastProps{Title: "Event created", Description: "Your event has been created.", Dismissible: true}),
				feedback.Toast(feedback.ToastProps{Title: "Success", Description: "Changes saved.", Variant: feedback.ToastSuccess, Dismissible: true}),
				feedback.Toast(feedback.ToastProps{Title: "Error", Description: "Something went wrong.", Variant: feedback.ToastError, Dismissible: true}),
				feedback.Toast(feedback.ToastProps{Title: "Warning", Description: "This action is irreversible.", Variant: feedback.ToastWarning, Dismissible: true}),
				feedback.Toast(feedback.ToastProps{Title: "Info", Description: "New updates are available.", Variant: feedback.ToastInfo, Dismissible: true}),
			)),
			example("Interactive — click to trigger (auto-dismisses after 5s)", h.Div(
				h.Class("flex flex-wrap gap-2"),
				toastTriggerButton("toast-tmpl-default", "Default"),
				toastTriggerButton("toast-tmpl-success", "Success"),
				toastTriggerButton("toast-tmpl-error", "Error"),
				toastTriggerButton("toast-tmpl-warning", "Warning"),
				toastTriggerButton("toast-tmpl-info", "Info"),
				toastTemplate("toast-tmpl-default", feedback.ToastDefault, "Event created", "Your event has been created."),
				toastTemplate("toast-tmpl-success", feedback.ToastSuccess, "Success", "Changes saved successfully."),
				toastTemplate("toast-tmpl-error", feedback.ToastError, "Error", "Something went wrong."),
				toastTemplate("toast-tmpl-warning", feedback.ToastWarning, "Warning", "This action is irreversible."),
				toastTemplate("toast-tmpl-info", feedback.ToastInfo, "Info", "New updates are available."),
			)),
		),

		// ── Popover ─────────────────────────────────────
		section("popover", "Popover",
			example("Default (left-aligned)", core.Popover.Root(core.PopoverRootProps{},
				core.Popover.Trigger(core.PopoverTriggerProps{Class: popoverBtnClass},
					g.Text("Open popover"),
				),
				core.Popover.Content(core.PopoverContentProps{},
					h.P(h.Class("p-4"),
						h.Span(h.Class("block text-sm font-medium mb-1"), g.Text("Popover title")),
						h.Span(h.Class("text-sm text-muted-foreground"), g.Text("This is a generic floating panel. It closes when you click outside.")),
					),
				),
			)),
			example("Right-aligned", core.Popover.Root(core.PopoverRootProps{},
				core.Popover.Trigger(core.PopoverTriggerProps{Class: popoverBtnClass},
					g.Text("Right-aligned"),
				),
				core.Popover.Content(core.PopoverContentProps{Align: "right"},
					h.P(h.Class("p-4 text-sm text-muted-foreground"), g.Text("Panel anchored to the right edge of the trigger.")),
				),
			)),
			example("Custom width", core.Popover.Root(core.PopoverRootProps{},
				core.Popover.Trigger(core.PopoverTriggerProps{Class: popoverBtnClass},
					g.Text("Narrow popover"),
				),
				core.Popover.Content(core.PopoverContentProps{Width: "w-48"},
					h.P(h.Class("p-4 text-sm text-muted-foreground"), g.Text("w-48 panel.")),
				),
			)),
		),

		// ── Tooltip ─────────────────────────────────────
		section("tooltip", "Tooltip",
			example("Hover or focus for tooltip", tooltip.Root(
				tooltip.Trigger(
					core.Button(core.ButtonProps{
						Label:   "Focus me",
						Variant: core.VariantOutline,
						Extra:   tooltip.TriggerAttrs("example-tooltip"),
					}),
				),
				tooltip.Content("example-tooltip", g.Text("This is a tooltip")),
			)),
			example("Click-activated (touch-friendly)", tooltip.ClickRoot(
				tooltip.ClickTrigger(
					g.Attr("aria-describedby", "click-tooltip"),
					core.Icon(iconrender.Props("lucide-icons", "circle-help", core.IconProps{
						Size:  "size-5",
						Label: "Help",
					})),
				),
				tooltip.ClickContent("click-tooltip", g.Text("Click or tap to reveal this tooltip")),
			)),
		),
	)
}

// popoverBtnClass applies outline-button styling to a <summary> trigger so it
// looks like a button without nesting an invalid <button> inside <summary>.
const popoverBtnClass = "gap-2 rounded-md border bg-background text-foreground shadow-xs hover:bg-accent hover:text-accent-foreground h-9 px-4 py-2 text-sm font-medium transition-all focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] outline-none"

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

// toastTriggerButton renders a button that clones a <template> toast into #toast-container.
func toastTriggerButton(templateID, label string) g.Node {
	return core.Button(core.ButtonProps{
		Label:   label,
		Variant: core.VariantOutline,
		Size:    core.SizeSm,
		Extra:   []g.Node{g.Attr("data-toast-trigger", templateID)},
	})
}

// toastTemplate renders a hidden <template> containing a toast for JS cloning.
// Position "" → container mode; JS injects the cloned node into #toast-container.
func toastTemplate(id string, variant feedback.ToastVariant, title, desc string) g.Node {
	return g.El("template", h.ID(id),
		feedback.Toast(feedback.ToastProps{
			Title:       title,
			Description: desc,
			Variant:     variant,
			Dismissible: true,
		}),
	)
}
