package config

import "github.com/go-playground/validator/v10"

// NavbarVisibility controls whether an item renders for guest, user, or all states.
type NavbarVisibility string

const (
	VisibilityAll   NavbarVisibility = "all"
	VisibilityGuest NavbarVisibility = "guest"
	VisibilityUser  NavbarVisibility = "user"
)

// NavbarSectionAlign controls where a section sits in the desktop and mobile layout.
type NavbarSectionAlign string

const (
	AlignStart NavbarSectionAlign = "start"
	AlignEnd   NavbarSectionAlign = "end"
)

// NavbarBrand configures the logo/wordmark shown at the start of the nav bar.
type NavbarBrand struct {
	Label    string `koanf:"label"`
	Href     string `koanf:"href"`
	LogoPath string `koanf:"logo_path"`
}

// NavHiddenField renders a hidden input inside a form-backed nav item.
type NavHiddenField struct {
	Name  string `koanf:"name" validate:"required"`
	Value string `koanf:"value"`
}

// NavConfig is the app-owned declarative nav configuration loaded from nav.yaml.
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
