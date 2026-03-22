package userpartial

import (
	"strings"
	"testing"

	"github.com/y-goweb/componentry/testutil"
)

func TestUserRowRendersActionsAndRoleBadge(t *testing.T) {
	got := testutil.RenderNode(t, UserRow(UserRowProps{User: partialTestUser}))

	wantSnippets := []string{
		`id="user-11111111-1111-1111-1111-111111111111"`,
		`Ada Lovelace`,
		`ada@example.com`,
		`admin`,
		`2026-03-20`,
		`hx-get="/users/11111111-1111-1111-1111-111111111111/edit"`,
		`hx-delete="/users/11111111-1111-1111-1111-111111111111"`,
		`hx-indicator="#users-loading"`,
		`hx-confirm="Delete Ada Lovelace?"`,
		`text-destructive hover:bg-destructive/10`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered row missing %q:\n%s", want, got)
		}
	}
}
