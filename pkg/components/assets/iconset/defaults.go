package iconset

import componenticons "starter/pkg/components/icons"

// Catalog bundles a sprite-file set with semantic icon bindings.
type Catalog struct {
	Sprites map[string]string
	Icons   map[componenticons.Key]componenticons.Ref
}

// Default is the built-in sprite and semantic icon catalog used by pkg/components.
var Default = Catalog{
	Sprites: map[string]string{
		"lucide-icons": "img/svg/lucide-icons.svg",
		"theme-icons":  "img/svg/theme-icons.svg",
	},
	Icons: map[componenticons.Key]componenticons.Ref{
		componenticons.ChevronDown:  {Sprite: "lucide-icons", ID: "chevron-down"},
		componenticons.ChevronLeft:  {Sprite: "lucide-icons", ID: "chevron-left"},
		componenticons.ChevronRight: {Sprite: "lucide-icons", ID: "chevron-right"},
		componenticons.ChevronsUp:   {Sprite: "lucide-icons", ID: "chevrons-up"},
		componenticons.Close:        {Sprite: "lucide-icons", ID: "x"},
		componenticons.ThemeLight:   {Sprite: "theme-icons", ID: "sun"},
		componenticons.ThemeDark:    {Sprite: "theme-icons", ID: "moon"},
		componenticons.ThemeSystem:  {Sprite: "theme-icons", ID: "monitor"},
	},
}
