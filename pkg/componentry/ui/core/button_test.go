package core

import (
	"strings"
	"testing"

	testutil "github.com/y-goweb/componentry/testutil"
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

func TestButtonSupportsDestructiveGhostVariant(t *testing.T) {
	got := testutil.RenderNode(t, Button(ButtonProps{
		Label:   "Delete",
		Variant: VariantDestructiveGhost,
	}))

	if !strings.Contains(got, `text-destructive`) || !strings.Contains(got, `hover:bg-destructive/10`) {
		t.Fatalf("Button() output missing destructive ghost styling: %s", got)
	}
}

func TestButtonLinkVariantUsesQuietTertiaryStyling(t *testing.T) {
	got := testutil.RenderNode(t, Button(ButtonProps{
		Label:   "Learn more",
		Variant: VariantLink,
	}))

	if !strings.Contains(got, `text-foreground`) || strings.Contains(got, `text-primary`) {
		t.Fatalf("Button() link output should stay quiet by default: %s", got)
	}
}
