package config

// NavbarVisibility controls whether an item renders for guest, user, or all states.
type NavbarVisibility string

// NavbarSectionAlign controls where a section sits in the desktop and mobile layout.
type NavbarSectionAlign string

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
