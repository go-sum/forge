package core

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
)

func renderAvatarNode(t *testing.T, node g.Node) string {
	t.Helper()

	var buf strings.Builder
	if err := node.Render(&buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return buf.String()
}

func TestAvatarImageRendersImgOnly(t *testing.T) {
	got := renderAvatarNode(t, Avatar.Image("/avatar.png", "User avatar"))

	if !strings.Contains(got, `<img class="aspect-square h-full w-full object-cover" src="/avatar.png" alt="User avatar">`) {
		t.Fatalf("Image() output missing img element: %s", got)
	}
	if strings.Contains(got, `bg-muted`) {
		t.Fatalf("Image() output unexpectedly rendered fallback markup: %s", got)
	}
}

func TestAvatarFallbackRendersPlaceholderOnly(t *testing.T) {
	got := renderAvatarNode(t, Avatar.Fallback(g.Text("AB")))

	if !strings.Contains(got, `bg-muted`) {
		t.Fatalf("Fallback() output missing placeholder styling: %s", got)
	}
	if strings.Contains(got, `<img`) {
		t.Fatalf("Fallback() output unexpectedly rendered an image: %s", got)
	}
}
