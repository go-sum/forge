package iconset

import icons "github.com/go-sum/componentry/icons"

// Catalog bundles a sprite-file set with semantic icon bindings.
type Catalog struct {
	Sprites map[string]string
	Icons   map[icons.Key]icons.Ref
}

// Default is the built-in sprite and semantic icon catalog used by pkg/components.
var Default = Catalog{
	Sprites: map[string]string{
		"lucide-icons": "img/svg/lucide-icons.svg",
		"theme-icons":  "img/svg/theme-icons.svg",
	},
	Icons: map[icons.Key]icons.Ref{
		icons.ChevronDown:  {Sprite: "lucide-icons", ID: "chevron-down"},
		icons.ChevronLeft:  {Sprite: "lucide-icons", ID: "chevron-left"},
		icons.ChevronRight: {Sprite: "lucide-icons", ID: "chevron-right"},
		icons.ChevronsUp:   {Sprite: "lucide-icons", ID: "chevrons-up"},
		icons.Close:        {Sprite: "lucide-icons", ID: "x"},
		icons.ThemeLight:   {Sprite: "theme-icons", ID: "sun"},
		icons.ThemeDark:    {Sprite: "theme-icons", ID: "moon"},
		icons.ThemeSystem:  {Sprite: "theme-icons", ID: "monitor"},
	},
}
