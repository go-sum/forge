package examples

import (
	"strings"
	"testing"
)

func TestPageRendersComponentShowcase(t *testing.T) {
	var buf strings.Builder
	if err := Page().Render(&buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	got := buf.String()
	checks := []string{"Component Examples", "Buttons", "Form Fields", "data-tabs", "HTMX Patterns", "Progressive Tiers", "Destructive Ghost"}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Page() output missing %q in %s", check, got)
		}
	}
}
