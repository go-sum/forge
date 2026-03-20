package dialog

import (
	"strings"
	"testing"

	testutil "starter/pkg/components/testutil"
	g "maragu.dev/gomponents"
)

func TestContentUsesGeneratedLabelAndDescriptionIDs(t *testing.T) {
	got := testutil.RenderNode(t, Content("confirm-dialog",
		Header(
			Title("confirm-dialog", g.Text("Confirm")),
			Description("confirm-dialog", g.Text("Are you sure?")),
		),
	))

	if !strings.Contains(got, ` aria-labelledby="confirm-dialog-title"`) {
		t.Fatalf("Content() output missing aria-labelledby: %s", got)
	}
	if !strings.Contains(got, ` aria-describedby="confirm-dialog-description"`) {
		t.Fatalf("Content() output missing aria-describedby: %s", got)
	}
	if !strings.Contains(got, ` id="confirm-dialog-title"`) {
		t.Fatalf("Title() output missing generated title id: %s", got)
	}
	if !strings.Contains(got, ` id="confirm-dialog-description"`) {
		t.Fatalf("Description() output missing generated description id: %s", got)
	}
}
