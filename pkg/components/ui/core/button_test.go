package core

import (
	"strings"
	"testing"

	testutil "starter/pkg/components/testutil"
)

func TestButtonDisabledLinkOmitsHrefAndMarksAriaDisabled(t *testing.T) {
	got := testutil.RenderNode(t, Button(ButtonProps{
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
