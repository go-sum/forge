package feedback

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	testutil "github.com/go-sum/componentry/testutil"
)

func TestAlertListMapsErrorTypeToDestructiveVariant(t *testing.T) {
	got := testutil.RenderNode(t, Alert.List([]string{"error"}, []string{"boom"}))

	if !strings.Contains(got, "text-destructive") {
		t.Fatalf("Alert.List() output missing destructive styling for error type: %s", got)
	}
}

func TestAlertRootWithIconRendersIconColumnAndWrapsChildren(t *testing.T) {
	icon := h.Span(g.Text("icon"))
	got := testutil.RenderNode(t, Alert.Root(
		AlertProps{Icon: icon},
		Alert.Title(g.Text("Heads up")),
	))

	if !strings.Contains(got, "data-alert-icon") {
		t.Fatalf("Alert.Root with Icon missing data-alert-icon wrapper: %s", got)
	}
	if !strings.Contains(got, "grid-cols-[auto_1fr]") {
		t.Fatalf("Alert.Root with Icon missing two-column grid layout: %s", got)
	}
	if !strings.Contains(got, "Heads up") {
		t.Fatalf("Alert.Root with Icon missing child content: %s", got)
	}
}

func TestAlertRootWithoutIconUsesDefaultLayout(t *testing.T) {
	got := testutil.RenderNode(t, Alert.Root(
		AlertProps{},
		Alert.Description(g.Text("all good")),
	))

	if strings.Contains(got, "data-alert-icon") {
		t.Fatalf("Alert.Root without Icon should not render icon wrapper: %s", got)
	}
	if strings.Contains(got, "grid-cols-[auto_1fr]") {
		t.Fatalf("Alert.Root without Icon should not use two-column layout: %s", got)
	}
}
