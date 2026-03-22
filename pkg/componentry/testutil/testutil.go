// Package testutil provides shared test helpers for the pkg/components tree.
package testutil

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
)

// RenderNode renders node to a string, failing the test on any render error.
func RenderNode(t *testing.T, node g.Node) string {
	t.Helper()
	var buf strings.Builder
	if err := node.Render(&buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return buf.String()
}
