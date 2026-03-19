package render

import (
	"testing"

	componentassets "starter/pkg/components/assets"
	componenticons "starter/pkg/components/icons"
	"starter/pkg/components/ui/core"
)

func TestProps(t *testing.T) {
	componentassets.SetPathFunc(func(rel string) string { return "/public/" + rel + "?v=test" })
	componentassets.RegisterSprite("test-lucide-icons", "img/svg/lucide-icons.svg")

	got := Props("test-lucide-icons", "chevron-down", core.IconProps{Size: "size-4"})
	if got.Src != "/public/img/svg/lucide-icons.svg?v=test" {
		t.Fatalf("Props().Src = %q", got.Src)
	}
	if got.ID != "chevron-down" {
		t.Fatalf("Props().ID = %q", got.ID)
	}
}

func TestPropsForRegistry(t *testing.T) {
	componentassets.SetPathFunc(func(rel string) string { return "/public/" + rel + "?v=test" })
	componentassets.RegisterSprite("test-hero-icons", "img/svg/hero-icons.svg")

	r := componenticons.NewRegistry()
	r.Register(componenticons.ChevronDown, componenticons.Ref{
		Sprite: "test-hero-icons",
		ID:     "chevron-down-solid",
	})

	got := PropsForRegistry(r, componenticons.ChevronDown, core.IconProps{Size: "size-4"})
	if got.Src != "/public/img/svg/hero-icons.svg?v=test" {
		t.Fatalf("PropsForRegistry().Src = %q", got.Src)
	}
	if got.ID != "chevron-down-solid" {
		t.Fatalf("PropsForRegistry().ID = %q", got.ID)
	}
}

func TestPropsForRegistries(t *testing.T) {
	assetRegistry := componentassets.NewRegistry()
	assetRegistry.SetPathFunc(func(rel string) string { return "/assets/" + rel })
	assetRegistry.RegisterSprite("test-custom-icons", "icons/custom.svg")

	iconRegistry := componenticons.NewRegistry()
	iconRegistry.Register(componenticons.ChevronRight, componenticons.Ref{
		Sprite: "test-custom-icons",
		ID:     "chevron-right",
	})

	got := PropsForRegistries(assetRegistry, iconRegistry, componenticons.ChevronRight, core.IconProps{Size: "size-4"})
	if got.Src != "/assets/icons/custom.svg" {
		t.Fatalf("PropsForRegistries().Src = %q", got.Src)
	}
	if got.ID != "chevron-right" {
		t.Fatalf("PropsForRegistries().ID = %q", got.ID)
	}
}

func TestPropsForRegistryUnknown(t *testing.T) {
	r := componenticons.NewRegistry()

	got := PropsForRegistry(r, componenticons.ChevronsUp, core.IconProps{Size: "size-4"})
	if got.Src != "" || got.ID != "" {
		t.Fatalf("PropsForRegistry() = %#v, want empty Src and ID", got)
	}
}
