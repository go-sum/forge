package feedback

import (
	"strings"
	"testing"

	testutil "github.com/go-sum/componentry/testutil"
)

func TestToastErrorUsesAlertAnnouncementSemantics(t *testing.T) {
	got := testutil.RenderNode(t, Toast(ToastProps{
		Title:       "Error",
		Description: "Save failed",
		Variant:     ToastError,
		Dismissible: true,
	}))

	if !strings.Contains(got, ` role="alert"`) {
		t.Fatalf("Toast() output missing alert role for error variant: %s", got)
	}
	if !strings.Contains(got, ` aria-live="assertive"`) {
		t.Fatalf("Toast() output missing assertive aria-live for error variant: %s", got)
	}
	if !strings.Contains(got, ` aria-label="Dismiss"`) {
		t.Fatalf("Toast() output missing dismiss button label: %s", got)
	}
}

func TestToastContainerModeUsesRelativeRootForDismissButton(t *testing.T) {
	got := testutil.RenderNode(t, Toast(ToastProps{
		Description: "Saved",
		Dismissible: true,
	}))

	if !strings.Contains(got, `class="relative rounded-lg border p-4 shadow-md`) {
		t.Fatalf("Toast() container output missing relative root class: %s", got)
	}
	if !strings.Contains(got, `absolute top-2 right-2 opacity-50 hover:opacity-100 transition-opacity`) || !strings.Contains(got, `focus-visible:ring-[3px]`) {
		t.Fatalf("Toast() container output missing absolutely positioned dismiss button: %s", got)
	}
}

func TestToastFixedModePreservesFixedPositioning(t *testing.T) {
	got := testutil.RenderNode(t, Toast(ToastProps{
		Description: "Saved",
		Position:    PositionBottomRight,
	}))

	if !strings.Contains(got, `class="fixed z-50 max-w-sm bottom-4 right-4 relative rounded-lg border p-4 shadow-md`) {
		t.Fatalf("Toast() fixed output missing fixed positioning classes: %s", got)
	}
}

func TestToastInfoUsesSemanticPrimaryTokens(t *testing.T) {
	got := testutil.RenderNode(t, Toast(ToastProps{
		Description: "Heads up",
		Variant:     ToastInfo,
	}))

	if !strings.Contains(got, `border-primary/30 bg-primary/20 text-primary`) {
		t.Fatalf("Toast() info output missing semantic token classes: %s", got)
	}
	if strings.Contains(got, `blue-`) {
		t.Fatalf("Toast() info output unexpectedly used raw blue palette classes: %s", got)
	}
}
