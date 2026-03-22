package adapters

import (
	uilayout "github.com/go-sum/componentry/ui/layout"
	"github.com/go-sum/forge/config"
)

// ToComponentryNavConfig converts forge's app-owned nav config into the
// componentry nav schema consumed by the shared UI shell.
func ToComponentryNavConfig(cfg config.NavConfig) uilayout.NavConfig {
	return uilayout.NavConfig{
		Brand: uilayout.NavbarBrand{
			Label:    cfg.Brand.Label,
			Href:     cfg.Brand.Href,
			LogoPath: cfg.Brand.LogoPath,
		},
		Sections: toComponentryNavSections(cfg.Sections),
	}
}

func toComponentryNavSections(sections []config.NavSection) []uilayout.NavSection {
	out := make([]uilayout.NavSection, 0, len(sections))
	for _, section := range sections {
		out = append(out, uilayout.NavSection{
			Label: section.Label,
			Align: uilayout.NavbarSectionAlign(section.Align),
			Items: toComponentryNavItems(section.Items),
		})
	}
	return out
}

func toComponentryNavItems(items []config.NavItem) []uilayout.NavItem {
	out := make([]uilayout.NavItem, 0, len(items))
	for _, item := range items {
		out = append(out, uilayout.NavItem{
			Type:         item.Type,
			Slot:         item.Slot,
			Visibility:   uilayout.NavbarVisibility(item.Visibility),
			Label:        item.Label,
			Href:         item.Href,
			Action:       item.Action,
			Method:       item.Method,
			Icon:         item.Icon,
			MatchPrefix:  item.MatchPrefix,
			HiddenFields: toComponentryHiddenFields(item.HiddenFields),
			Items:        toComponentryNavItems(item.Items),
		})
	}
	return out
}

func toComponentryHiddenFields(fields []config.NavHiddenField) []uilayout.NavHiddenField {
	out := make([]uilayout.NavHiddenField, 0, len(fields))
	for _, field := range fields {
		out = append(out, uilayout.NavHiddenField{
			Name:  field.Name,
			Value: field.Value,
		})
	}
	return out
}
