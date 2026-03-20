package form

import (
	"strings"
	"testing"

	testutil "starter/pkg/components/testutil"
)

func TestSelectRendersOptGroups(t *testing.T) {
	got := testutil.RenderNode(t, Select(SelectProps{
		ID:       "role",
		Name:     "role",
		Selected: "admin",
		Groups: []OptGroup{
			{Label: "Admin roles", Options: []Option{{Value: "admin", Label: "Admin"}}},
			{Label: "Other roles", Options: []Option{{Value: "viewer", Label: "Viewer"}}},
		},
	}))

	if !strings.Contains(got, `<optgroup label="Admin roles">`) {
		t.Fatalf("Select() output missing optgroup label in %s", got)
	}
	if !strings.Contains(got, `<option selected value="admin">Admin</option>`) {
		t.Fatalf("Select() output missing selected admin option in %s", got)
	}
}

func TestSelectMultipleRendersMultipleAttribute(t *testing.T) {
	got := testutil.RenderNode(t, Select(SelectProps{
		ID:       "roles",
		Name:     "roles",
		Multiple: true,
		Options: []Option{
			{Value: "admin", Label: "Admin"},
			{Value: "editor", Label: "Editor"},
		},
		SelectedValues: []string{"editor"},
	}))

	if !strings.Contains(got, ` multiple=""`) {
		t.Fatalf("Select() output missing multiple attribute: %s", got)
	}
	if !strings.Contains(got, `selected value="editor">Editor</option>`) {
		t.Fatalf("Select() output missing selected option for multiple select: %s", got)
	}
}
