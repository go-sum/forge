package tooltip

import (
	"strings"
	"testing"

	testutil "github.com/y-goweb/componentry/testutil"
	g "maragu.dev/gomponents"
)

func TestClickVariantRendersDetailsWithPopoverHook(t *testing.T) {
	got := testutil.RenderNode(t, ClickRoot(
		ClickTrigger(g.Attr("aria-describedby", "click-tip"), g.Text("?")),
		ClickContent("click-tip", g.Text("Help text")),
	))

	checks := []string{
		`<details`,
		`data-popover=""`,
		`<summary`,
		`list-none`,
		`aria-describedby="click-tip"`,
		`id="click-tip"`,
		`role="tooltip"`,
		`Help text`,
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("ClickRoot() missing %q in:\n%s", want, got)
		}
	}
	// Click variant must NOT use hover-only visibility classes.
	if strings.Contains(got, `group-hover:block`) {
		t.Errorf("ClickRoot() should not contain group-hover:block in: %s", got)
	}
}

func TestContentAndTriggerAttrsExposeTooltipSemantics(t *testing.T) {
	got := testutil.RenderNode(t, Root(
		Trigger(g.El("button", append([]g.Node{g.Attr("type", "button")}, TriggerAttrs("help-tip")...)...)),
		Content("help-tip", g.Text("Helpful text")),
	))

	if !strings.Contains(got, ` aria-describedby="help-tip"`) {
		t.Fatalf("TriggerAttrs() output missing aria-describedby: %s", got)
	}
	if !strings.Contains(got, ` id="help-tip"`) {
		t.Fatalf("Content() output missing tooltip id: %s", got)
	}
	if !strings.Contains(got, ` role="tooltip"`) {
		t.Fatalf("Content() output missing tooltip role: %s", got)
	}
	if !strings.Contains(got, `group-focus-within:block`) {
		t.Fatalf("Content() output missing focus-within visibility class: %s", got)
	}
}
