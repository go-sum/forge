package install

import (
	componenticonset "github.com/go-sum/componentry/assets/iconset"
	componenticons "github.com/go-sum/componentry/icons"
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
