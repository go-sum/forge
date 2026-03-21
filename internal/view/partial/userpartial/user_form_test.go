package userpartial

import (
	"strings"
	"testing"
	"time"

	"starter/internal/model"
	"starter/internal/view"
	"starter/pkg/components/testutil"

	"github.com/google/uuid"
)

var partialTestUser = model.User{
	ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	Email:       "ada@example.com",
	DisplayName: "Ada Lovelace",
	Role:        "admin",
	CreatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
}

func TestUserEditFormRendersValuesErrorsAndHTMXAttrs(t *testing.T) {
	got := testutil.RenderNode(t, UserEditForm(view.Request{
		CSRFToken: "csrf-token",
	}, UserFormData{
		User: partialTestUser,
		Values: model.UpdateUserInput{
			Email:       "grace@example.com",
			DisplayName: "Grace Hopper",
			Role:        "user",
		},
		Errors: map[string][]string{
			"Email": {"Email already in use."},
			"_":     {"Save failed."},
		},
	}))

	wantSnippets := []string{
		`id="user-11111111-1111-1111-1111-111111111111"`,
		`hx-put="/users/11111111-1111-1111-1111-111111111111"`,
		`hx-indicator="#users-loading"`,
		`sm:grid-cols-2`,
		`value="csrf-token"`,
		`value="grace@example.com"`,
		`value="Grace Hopper"`,
		`Email already in use.`,
		`Save failed.`,
		`hx-get="/users/11111111-1111-1111-1111-111111111111/row"`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered edit form missing %q:\n%s", want, got)
		}
	}
}
