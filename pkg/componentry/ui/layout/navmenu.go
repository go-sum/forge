package layout

import (
	componenticons "github.com/y-goweb/componentry/icons"

	"github.com/go-playground/validator/v10"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// NavConfig is the high-level declarative configuration consumed by NavMenu.
type NavConfig struct {
	Brand    NavbarBrand  `koanf:"brand"`
	Sections []NavSection `koanf:"sections" validate:"dive"`
}

// NavSection groups a list of declarative menu items.
type NavSection struct {
	Label string             `koanf:"label"`
	Align NavbarSectionAlign `koanf:"align" validate:"omitempty,oneof=start end"`
	Items []NavItem          `koanf:"items" validate:"min=1,dive"`
}

// NavItem describes one link, submenu, separator, form action, or named slot.
// Type is reserved for structural markers such as separator; Slot names
// caller-supplied dynamic content such as theme toggles or auth controls.
type NavItem struct {
	Type         string           `koanf:"type" validate:"omitempty,oneof=separator"`
	Slot         string           `koanf:"slot"`
	Visibility   NavbarVisibility `koanf:"visibility" validate:"omitempty,oneof=all guest user"`
	Label        string           `koanf:"label"`
	Href         string           `koanf:"href"`
	Action       string           `koanf:"action"`
	Method       string           `koanf:"method" validate:"omitempty,oneof=get post"`
	Icon         string           `koanf:"icon"`
	MatchPrefix  bool             `koanf:"match_prefix"`
	HiddenFields []NavHiddenField `koanf:"hidden_fields" validate:"omitempty,dive"`
	Items        []NavItem        `koanf:"items" validate:"omitempty,dive"`
}

// NavSlot provides caller-supplied content for a named slot in the config.
type NavSlot struct {
	Desktop g.Node
	Mobile  g.Node
}

// NavSlots indexes named config slots such as theme toggles or auth controls.
type NavSlots map[string]NavSlot

// NavMenuProps configures NavMenu, the high-level declarative menu wrapper.
type NavMenuProps struct {
	ID              string
	Config          NavConfig
	Slots           NavSlots
	CurrentPath     string
	IsAuthenticated bool
}

// FormSlotProps configures a reusable form slot such as logout.
type FormSlotProps struct {
	Label        string
	Action       string
	Method       string
	Icon         componenticons.Key
	HiddenFields []NavHiddenField
}

// NavMenu renders a declarative nav config through the lower-level Navbar shell.
func NavMenu(p NavMenuProps) g.Node {
	return Navbar(NavbarProps{
		ID:              p.ID,
		Brand:           p.Config.Brand,
		Sections:        buildSections(p.Config.Sections, p.Slots),
		CurrentPath:     p.CurrentPath,
		IsAuthenticated: p.IsAuthenticated,
	})
}

// TextSlot renders non-interactive text in desktop and mobile-specific styles.
func TextSlot(text string) NavSlot {
	if text == "" {
		return NavSlot{}
	}
	return NavSlot{
		Desktop: h.Span(h.Class("text-sm text-muted-foreground"), g.Text(text)),
		Mobile:  h.Span(h.Class("block px-4 py-4 text-sm font-medium text-foreground"), g.Text(text)),
	}
}

// ControlSlot renders a desktop control directly and wraps it in a labeled mobile row.
func ControlSlot(label string, control g.Node) NavSlot {
	if control == nil {
		return NavSlot{}
	}
	return NavSlot{
		Desktop: control,
		Mobile: h.Div(
			h.Class("flex items-center justify-between px-4 py-4 transition-colors hover:bg-accent/60"),
			h.Span(h.Class("text-sm text-muted-foreground"), g.Text(label)),
			control,
		),
	}
}

// FormSlot renders a form action using the shared NavForm component styling.
func FormSlot(p FormSlotProps) NavSlot {
	form := NavForm{
		Label:        p.Label,
		Action:       p.Action,
		Method:       p.Method,
		Icon:         p.Icon,
		HiddenFields: p.HiddenFields,
	}
	return NavSlot{
		Desktop: form.render(navbarItemContext{viewport: viewportDesktop}),
		Mobile:  form.render(navbarItemContext{viewport: viewportMobile}),
	}
}

func buildSections(sections []NavSection, slots NavSlots) []NavbarSection {
	built := make([]NavbarSection, 0, len(sections))
	for _, section := range sections {
		items := buildItems(section.Items, slots)
		built = append(built, NavbarSection{
			Label: section.Label,
			Align: section.Align,
			Items: items,
		})
	}
	return built
}

func buildItems(items []NavItem, slots NavSlots) []NavbarItem {
	built := make([]NavbarItem, 0, len(items))
	for _, item := range items {
		node := buildItem(item, slots)
		if node != nil {
			built = append(built, node)
		}
	}
	return built
}

func buildItem(item NavItem, slots NavSlots) NavbarItem {
	if item.Type == "separator" {
		return NavSeparator{Visibility: item.Visibility}
	}
	if slotName := slotName(item); slotName != "" {
		slot, ok := slots[slotName]
		if !ok {
			return nil
		}
		return NavNode{Visibility: item.Visibility, Desktop: slot.Desktop, Mobile: slot.Mobile}
	}

	icon := componenticons.Key(item.Icon)
	if len(item.Items) > 0 {
		return NavGroup{
			Visibility:  item.Visibility,
			Label:       item.Label,
			Href:        item.Href,
			Icon:        icon,
			MatchPrefix: item.MatchPrefix,
			Items:       buildItems(item.Items, slots),
		}
	}
	if item.Action != "" {
		return NavForm{
			Visibility:   item.Visibility,
			Label:        item.Label,
			Action:       item.Action,
			Method:       item.Method,
			Icon:         icon,
			HiddenFields: item.HiddenFields,
		}
	}
	if item.Href != "" {
		return NavLink{
			Visibility:  item.Visibility,
			Label:       item.Label,
			Href:        item.Href,
			Icon:        icon,
			MatchPrefix: item.MatchPrefix,
		}
	}
	if item.Label != "" {
		return NavText{Visibility: item.Visibility, Text: item.Label, Icon: icon}
	}
	return nil
}

func slotName(item NavItem) string {
	return item.Slot
}

// RegisterNavValidations registers the declarative nav schema rules on v.
func RegisterNavValidations(v *validator.Validate) {
	v.RegisterStructValidation(navItemStructValidation, NavItem{})
}

func navItemStructValidation(sl validator.StructLevel) {
	item := sl.Current().Interface().(NavItem)

	if item.Type == "separator" {
		reportIfSet(sl, item.Slot, "Slot", "slot", "separator_only")
		reportIfSet(sl, item.Label, "Label", "label", "separator_only")
		reportIfSet(sl, item.Href, "Href", "href", "separator_only")
		reportIfSet(sl, item.Action, "Action", "action", "separator_only")
		reportIfSet(sl, item.Method, "Method", "method", "separator_only")
		reportIfSet(sl, item.Icon, "Icon", "icon", "separator_only")
		reportIfTrue(sl, item.MatchPrefix, "MatchPrefix", "match_prefix", "separator_only")
		reportIfLen(sl, len(item.HiddenFields), "HiddenFields", "hidden_fields", "separator_only")
		reportIfLen(sl, len(item.Items), "Items", "items", "separator_only")
		return
	}

	if item.MatchPrefix && item.Href == "" {
		sl.ReportError(item.MatchPrefix, "MatchPrefix", "match_prefix", "requires_href", "")
	}
	if item.Method != "" && item.Action == "" {
		sl.ReportError(item.Method, "Method", "method", "requires_action", "")
	}
	if len(item.HiddenFields) > 0 && item.Action == "" {
		sl.ReportError(item.HiddenFields, "HiddenFields", "hidden_fields", "requires_action", "")
	}

	if item.Slot != "" {
		reportIfSet(sl, item.Href, "Href", "href", "slot_conflict")
		reportIfSet(sl, item.Action, "Action", "action", "slot_conflict")
		reportIfSet(sl, item.Method, "Method", "method", "slot_conflict")
		reportIfTrue(sl, item.MatchPrefix, "MatchPrefix", "match_prefix", "slot_conflict")
		reportIfLen(sl, len(item.HiddenFields), "HiddenFields", "hidden_fields", "slot_conflict")
		reportIfLen(sl, len(item.Items), "Items", "items", "slot_conflict")
		return
	}

	hasHref := item.Href != ""
	hasAction := item.Action != ""
	hasItems := len(item.Items) > 0

	if hasHref && hasAction {
		sl.ReportError(item.Action, "Action", "action", "conflicts_with_href", "")
	}
	if hasAction && hasItems {
		sl.ReportError(item.Action, "Action", "action", "conflicts_with_items", "")
	}

	if (hasHref || hasAction || hasItems) && item.Label == "" {
		sl.ReportError(item.Label, "Label", "label", "required_for_item", "")
	}

	if !hasHref && !hasAction && !hasItems && item.Label == "" {
		sl.ReportError(item.Label, "Label", "label", "required", "")
	}
}

func reportIfSet(sl validator.StructLevel, value string, fieldName, jsonName, tag string) {
	if value != "" {
		sl.ReportError(value, fieldName, jsonName, tag, "")
	}
}

func reportIfTrue(sl validator.StructLevel, value bool, fieldName, jsonName, tag string) {
	if value {
		sl.ReportError(value, fieldName, jsonName, tag, "")
	}
}

func reportIfLen(sl validator.StructLevel, n int, fieldName, jsonName, tag string) {
	if n > 0 {
		sl.ReportError(n, fieldName, jsonName, tag, "")
	}
}
