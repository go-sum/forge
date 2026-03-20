package tabs

import (
	"strings"
	"testing"

	testutil "starter/pkg/components/testutil"
	g "maragu.dev/gomponents"
)

func TestTriggerAndContentExposeARIAWiring(t *testing.T) {
	got := testutil.RenderNode(t, Root("settings-tabs", "account",
		List(
			Trigger("settings-tabs", "account", true, g.Text("Account")),
		),
		Content("settings-tabs", "account", true, g.Text("Panel")),
	))

	if !strings.Contains(got, ` id="settings-tabs-tab-account"`) {
		t.Fatalf("Trigger() output missing generated tab id: %s", got)
	}
	if !strings.Contains(got, ` aria-controls="settings-tabs-panel-account"`) {
		t.Fatalf("Trigger() output missing aria-controls: %s", got)
	}
	if !strings.Contains(got, ` aria-labelledby="settings-tabs-tab-account"`) {
		t.Fatalf("Content() output missing aria-labelledby: %s", got)
	}
	if !strings.Contains(got, ` tabindex="0"`) {
		t.Fatalf("Tabs output missing initial roving tabindex: %s", got)
	}
}
