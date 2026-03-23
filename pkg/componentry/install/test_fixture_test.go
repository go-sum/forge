package install

import (
	"github.com/go-sum/componentry/assets/iconset"
	icons "github.com/go-sum/componentry/icons"
)

func iconsetFixture() iconset.Catalog {
	return iconset.Catalog{
		Sprites: map[string]string{
			"custom": "icons/custom.svg",
		},
		Icons: map[icons.Key]icons.Ref{
			icons.ChevronDown: {Sprite: "custom", ID: "chevron-down"},
		},
	}
}
