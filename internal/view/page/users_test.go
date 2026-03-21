package page

import (
	"strings"
	"testing"
	"time"

	"starter/internal/model"
	"starter/internal/view"
	"starter/pkg/components/patterns/pager"
	"starter/pkg/components/testutil"
)

func TestUserListRegionIsHTMXReplaceable(t *testing.T) {
	pg := pager.Pager{Page: 2, PerPage: 20, TotalItems: 45, TotalPages: 3}
	got := testutil.RenderNode(t, UserListRegion(UserListData{
		Users: []model.User{{
			DisplayName: "Ada Lovelace",
			Email:       "ada@example.com",
			Role:        "admin",
			CreatedAt:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		}},
		Pager: pg,
	}))

	wantSnippets := []string{
		`id="users-list-region"`,
		`id="users-table"`,
		`hx-target="#users-list-region"`,
		`hx-get="/users?page=1"`,
		`hx-get="/users?page=3"`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered region missing %q:\n%s", want, got)
		}
	}
}

func TestUserListPageRendersShellFromRequest(t *testing.T) {
	got := testutil.RenderNode(t, UserListPage(view.Request{
		CSRFToken:       "csrf-token",
		IsAuthenticated: true,
	}, UserListData{}))

	if !strings.Contains(got, "<html") || !strings.Contains(got, "Users") {
		t.Fatalf("rendered page = %q", got)
	}
}
