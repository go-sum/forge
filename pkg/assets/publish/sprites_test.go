package publish

import "testing"

func TestRegistrySpritePath(t *testing.T) {
	r := NewRegistry()
	r.RegisterSprite("lucide-icons", "img/svg/lucide-icons.svg")
	r.SetPathFunc(func(rel string) string { return "/public/" + rel + "?v=test" })

	got := r.SpritePath("lucide-icons")
	want := "/public/img/svg/lucide-icons.svg?v=test"
	if got != want {
		t.Fatalf("SpritePath() = %q, want %q", got, want)
	}
}

func TestRegistrySpritePathUnknown(t *testing.T) {
	r := NewRegistry()

	if got := r.SpritePath("missing"); got != "" {
		t.Fatalf("SpritePath() = %q, want empty string", got)
	}
}
