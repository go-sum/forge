package core

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
)

func renderNode(t *testing.T, node g.Node) string {
	t.Helper()

	var buf strings.Builder
	if err := node.Render(&buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return buf.String()
}

func TestButtonDisabledLinkOmitsHrefAndMarksAriaDisabled(t *testing.T) {
	got := renderNode(t, Button(Props{
		Label:    "Users",
		Href:     "/users",
		Disabled: true,
	}))

	if strings.Contains(got, ` href="/users"`) {
		t.Fatalf("Button() output kept href for disabled link: %s", got)
	}
	if !strings.Contains(got, ` aria-disabled="true"`) {
		t.Fatalf("Button() output missing aria-disabled for disabled link: %s", got)
	}
	if !strings.Contains(got, ` tabindex="-1"`) {
		t.Fatalf("Button() output missing tabindex for disabled link: %s", got)
	}
}
