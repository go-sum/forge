package install

import (
	"testing"

	assets "github.com/go-sum/componentry/assets"
	icons "github.com/go-sum/componentry/icons"
)

func TestNewBuildsIsolatedRegistries(t *testing.T) {
	regs := New(Config{
		PathFunc: func(rel string) string { return "/assets/" + rel },
		IconOverrides: map[icons.Key]icons.Ref{
			icons.ChevronDown: {Sprite: "custom", ID: "down"},
		},
		Catalog: iconsetFixture(),
	})

	if got := regs.Assets.SpritePath("custom"); got != "/assets/icons/custom.svg" {
		t.Fatalf("Assets.SpritePath() = %q, want %q", got, "/assets/icons/custom.svg")
	}
	if got, ok := regs.Icons.Resolve(icons.ChevronDown); !ok || got.ID != "down" {
		t.Fatalf("Icons.Resolve() = %#v, ok=%v", got, ok)
	}
}

func TestApplyDefaultUsesCurrentDefaultRegistries(t *testing.T) {
	oldAssets := assets.DefaultRegistry
	oldIcons := icons.Default
	assets.DefaultRegistry = assets.NewRegistry()
	icons.Default = icons.NewRegistry()
	t.Cleanup(func() {
		assets.DefaultRegistry = oldAssets
		icons.Default = oldIcons
	})

	ApplyDefault(Config{Catalog: iconsetFixture()})

	if got := assets.SpritePath("custom"); got == "" {
		t.Fatal("ApplyDefault() did not populate default asset registry")
	}
	if _, ok := icons.Resolve(icons.ChevronDown); !ok {
		t.Fatal("ApplyDefault() did not populate default icon registry")
	}
}
