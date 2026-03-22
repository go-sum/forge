package page

import (
	"strings"
	"testing"
	"time"

	"github.com/go-sum/componentry/patterns/pager"
	"github.com/go-sum/componentry/testutil"
	"github.com/go-sum/forge/config"
	"github.com/go-sum/forge/internal/model"
	"github.com/go-sum/forge/internal/view"
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
		`id="users-loading"`,
		`Updating users...`,
		` aria-label="pagination"`,
		`hx-target="#users-list-region"`,
		`hx-indicator="#users-loading"`,
		`hx-get="/users?page=1"`,
		`hx-get="/users?page=3"`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered region missing %q:\n%s", want, got)
		}
	}
}

func TestUserListRegionRendersEmptyStateWhenNoUsersExist(t *testing.T) {
	got := testutil.RenderNode(t, UserListRegion(UserListData{
		Pager: pager.Pager{Page: 1, PerPage: 20, TotalItems: 0, TotalPages: 0},
	}))

	wantSnippets := []string{
		`No users yet`,
		`User accounts will appear here once people register.`,
		`id="users-loading"`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered empty region missing %q:\n%s", want, got)
		}
	}
}

func TestUserListPageRendersShellFromRequest(t *testing.T) {
	got := testutil.RenderNode(t, UserListPage(view.Request{
		CurrentPath:     "/users",
		CSRFToken:       "csrf-token",
		IsAuthenticated: true,
		NavConfig: config.NavConfig{
			Brand: config.NavbarBrand{Label: "Starter", Href: "/"},
			Sections: []config.NavSection{{
				Items: []config.NavItem{{Label: "Users", Href: "/users"}},
			}},
		},
	}, UserListData{}))

	if !strings.Contains(got, "<html") || !strings.Contains(got, "Users") || !strings.Contains(got, `Manage account records with inline edits`) || !strings.Contains(got, `aria-current="page"`) {
		t.Fatalf("rendered page = %q", got)
	}
}
