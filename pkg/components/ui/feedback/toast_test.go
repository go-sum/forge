package feedback

import (
	"strings"
	"testing"

	testutil "starter/pkg/components/testutil"
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
	if !strings.Contains(got, `class="absolute top-2 right-2 opacity-50 hover:opacity-100 transition-opacity"`) {
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
