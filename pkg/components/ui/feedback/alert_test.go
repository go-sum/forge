package feedback

import (
	"strings"
	"testing"
)

func TestAlertListMapsErrorTypeToDestructiveVariant(t *testing.T) {
	got := renderNode(t, Alert.List([]string{"error"}, []string{"boom"}))

	if !strings.Contains(got, "text-destructive") {
		t.Fatalf("Alert.List() output missing destructive styling for error type: %s", got)
	}
}
