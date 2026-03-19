package feedback

import (
	"strings"
	"testing"
)

func TestToastErrorUsesAlertAnnouncementSemantics(t *testing.T) {
	got := renderNode(t, Toast(ToastProps{
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
