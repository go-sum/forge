package icons

import "testing"

func TestRegistryResolve(t *testing.T) {
	r := NewRegistry()
	r.Register(ChevronDown, Ref{Sprite: "hero-icons", ID: "chevron-down-solid"})

	got, ok := r.Resolve(ChevronDown)
	if !ok {
		t.Fatal("Resolve() returned ok=false, want true")
	}
	if got.Sprite != "hero-icons" || got.ID != "chevron-down-solid" {
		t.Fatalf("Resolve() = %#v, want hero-icons/chevron-down-solid", got)
	}
}

func TestRegistryResolveUnknown(t *testing.T) {
	r := NewRegistry()

	if got, ok := r.Resolve(ChevronsUp); ok {
		t.Fatalf("Resolve() = %#v, ok=true, want ok=false", got)
	}
}
