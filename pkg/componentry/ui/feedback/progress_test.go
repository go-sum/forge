package feedback

import (
	"strings"
	"testing"

	testutil "github.com/go-sum/componentry/testutil"
)

func TestProgressRendersLabelAndComputedPercentage(t *testing.T) {
	got := testutil.RenderNode(t, Progress(ProgressProps{
		ID:        "sync-progress",
		Label:     "Syncing",
		Value:     25,
		Max:       100,
		ShowValue: true,
		Size:      ProgressLg,
		Variant:   ProgressSuccess,
	}))

	checks := []string{` id="sync-progress"`, `>Syncing</span>`, `>25%</span>`, `progress-success`, ` max="100"`, ` value="25"`}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Progress() output missing %q in %s", check, got)
		}
	}
}
