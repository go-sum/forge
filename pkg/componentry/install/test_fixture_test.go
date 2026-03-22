package install

import (
	componenticonset "github.com/y-goweb/componentry/assets/iconset"
	componenticons "github.com/y-goweb/componentry/icons"
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
