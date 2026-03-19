package core

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
)

func TestPopoverRootRendersDetailsWithDataPopover(t *testing.T) {
	got := renderNode(t, Popover.Root(PopoverRootProps{ID: "info"},
		Popover.Trigger(PopoverTriggerProps{}, g.Text("Open")),
		Popover.Content(PopoverContentProps{}, g.Text("Hello")),
	))

	checks := []string{
		`<details`,
		`data-popover=""`,
		`id="info"`,
		`<summary`,
		`list-none`,
		`cursor-pointer`,
		`Hello`,
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("Popover.Root() missing %q in:\n%s", want, got)
		}
	}
}

func TestPopoverTriggerClassAppended(t *testing.T) {
	got := renderNode(t, Popover.Trigger(PopoverTriggerProps{
		Class: "rounded-md border px-3 py-2",
	}, g.Text("Click")))

	if !strings.Contains(got, "list-none") {
		t.Errorf("Popover.Trigger() missing base list-none in: %s", got)
	}
	if !strings.Contains(got, "rounded-md") {
		t.Errorf("Popover.Trigger() missing appended class in: %s", got)
	}
}

func TestPopoverContentAlignRight(t *testing.T) {
	got := renderNode(t, Popover.Content(PopoverContentProps{Align: "right"}, g.Text("x")))
	if !strings.Contains(got, "right-0") {
		t.Errorf("Popover.Content(right) missing right-0 in: %s", got)
	}
}

func TestPopoverContentAlignCenter(t *testing.T) {
	got := renderNode(t, Popover.Content(PopoverContentProps{Align: "center"}, g.Text("x")))
	if !strings.Contains(got, "left-1/2") || !strings.Contains(got, "-translate-x-1/2") {
		t.Errorf("Popover.Content(center) missing centering classes in: %s", got)
	}
}

func TestPopoverRootCustomClass(t *testing.T) {
	got := renderNode(t, Popover.Root(PopoverRootProps{Class: "relative inline-flex"}, g.Text("x")))
	if !strings.Contains(got, `class="relative inline-flex"`) {
		t.Errorf("Popover.Root() did not override class in: %s", got)
	}
}
