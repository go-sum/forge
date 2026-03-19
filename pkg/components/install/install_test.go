package install

import (
	"testing"

	componentassets "starter/pkg/components/assets"
	componenticons "starter/pkg/components/icons"
)

func TestNewBuildsIsolatedRegistries(t *testing.T) {
	regs := New(Config{
		PathFunc: func(rel string) string { return "/assets/" + rel },
		IconOverrides: map[componenticons.Key]componenticons.Ref{
			componenticons.ChevronDown: {Sprite: "custom", ID: "down"},
		},
		Catalog: componentassetsiconsetFixture(),
	})

	if got := regs.Assets.SpritePath("custom"); got != "/assets/icons/custom.svg" {
		t.Fatalf("Assets.SpritePath() = %q, want %q", got, "/assets/icons/custom.svg")
	}
	if got, ok := regs.Icons.Resolve(componenticons.ChevronDown); !ok || got.ID != "down" {
		t.Fatalf("Icons.Resolve() = %#v, ok=%v", got, ok)
	}
}

func TestApplyDefaultUsesCurrentDefaultRegistries(t *testing.T) {
	oldAssets := componentassets.Default
	oldIcons := componenticons.Default
	componentassets.Default = componentassets.NewRegistry()
	componenticons.Default = componenticons.NewRegistry()
	t.Cleanup(func() {
		componentassets.Default = oldAssets
		componenticons.Default = oldIcons
	})

	ApplyDefault(Config{Catalog: componentassetsiconsetFixture()})

	if got := componentassets.SpritePath("custom"); got == "" {
		t.Fatal("ApplyDefault() did not populate default asset registry")
	}
	if _, ok := componenticons.Resolve(componenticons.ChevronDown); !ok {
		t.Fatal("ApplyDefault() did not populate default icon registry")
	}
}
