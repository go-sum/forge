package iconset

import "testing"

func TestDefaultIconsUseRegisteredSprites(t *testing.T) {
	for key, ref := range Default.Icons {
		if _, ok := Default.Sprites[ref.Sprite]; !ok {
			t.Fatalf("default icon %q uses unknown sprite %q", key, ref.Sprite)
		}
		if ref.ID == "" {
			t.Fatalf("default icon %q has empty symbol id", key)
		}
	}
}
