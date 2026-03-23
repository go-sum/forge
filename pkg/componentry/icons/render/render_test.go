package render

import (
	"testing"

	assets "github.com/go-sum/componentry/assets"
	icons "github.com/go-sum/componentry/icons"
	"github.com/go-sum/componentry/ui/core"
)

func TestProps(t *testing.T) {
	assets.SetPathFunc(func(rel string) string { return "/public/" + rel + "?v=test" })
	assets.RegisterSprite("test-lucide-icons", "img/svg/lucide-icons.svg")

	got := Props("test-lucide-icons", "chevron-down", core.IconProps{Size: "size-4"})
	if got.Src != "/public/img/svg/lucide-icons.svg?v=test" {
		t.Fatalf("Props().Src = %q", got.Src)
	}
	if got.ID != "chevron-down" {
		t.Fatalf("Props().ID = %q", got.ID)
	}
}

func TestPropsForRegistry(t *testing.T) {
	assets.SetPathFunc(func(rel string) string { return "/public/" + rel + "?v=test" })
	assets.RegisterSprite("test-hero-icons", "img/svg/hero-icons.svg")

	r := icons.NewRegistry()
	r.Register(icons.ChevronDown, icons.Ref{
		Sprite: "test-hero-icons",
		ID:     "chevron-down-solid",
	})

	got := PropsForRegistry(r, icons.ChevronDown, core.IconProps{Size: "size-4"})
	if got.Src != "/public/img/svg/hero-icons.svg?v=test" {
		t.Fatalf("PropsForRegistry().Src = %q", got.Src)
	}
	if got.ID != "chevron-down-solid" {
		t.Fatalf("PropsForRegistry().ID = %q", got.ID)
	}
}

func TestPropsForRegistries(t *testing.T) {
	assetRegistry := assets.NewRegistry()
	assetRegistry.SetPathFunc(func(rel string) string { return "/assets/" + rel })
	assetRegistry.RegisterSprite("test-custom-icons", "icons/custom.svg")

	iconRegistry := icons.NewRegistry()
	iconRegistry.Register(icons.ChevronRight, icons.Ref{
		Sprite: "test-custom-icons",
		ID:     "chevron-right",
	})

	got := PropsForRegistries(assetRegistry, iconRegistry, icons.ChevronRight, core.IconProps{Size: "size-4"})
	if got.Src != "/assets/icons/custom.svg" {
		t.Fatalf("PropsForRegistries().Src = %q", got.Src)
	}
	if got.ID != "chevron-right" {
		t.Fatalf("PropsForRegistries().ID = %q", got.ID)
	}
}

func TestPropsForRegistryUnknown(t *testing.T) {
	r := icons.NewRegistry()

	got := PropsForRegistry(r, icons.ChevronsUp, core.IconProps{Size: "size-4"})
	if got.Src != "" || got.ID != "" {
		t.Fatalf("PropsForRegistry() = %#v, want empty Src and ID", got)
	}
}
