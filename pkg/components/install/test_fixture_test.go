package install

import (
	componenticonset "starter/pkg/components/assets/iconset"
	componenticons "starter/pkg/components/icons"
)

func componentassetsiconsetFixture() componenticonset.Catalog {
	return componenticonset.Catalog{
		Sprites: map[string]string{
			"custom": "icons/custom.svg",
		},
		Icons: map[componenticons.Key]componenticons.Ref{
			componenticons.ChevronDown: {Sprite: "custom", ID: "chevron-down"},
		},
	}
}
