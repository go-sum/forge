package tabs

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

func TestTriggerAndContentExposeARIAWiring(t *testing.T) {
	got := renderNode(t, Root("settings-tabs", "account",
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
