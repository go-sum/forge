package data

import (
	"strings"
	"testing"

	g "maragu.dev/gomponents"
	testutil "github.com/go-sum/componentry/testutil"
)

func TestCardRendersStructuredSections(t *testing.T) {
	got := testutil.RenderNode(t, Card.Root(
		Card.Header(Card.Title(g.Text("Users")), Card.Description(g.Text("Manage user accounts."))),
		Card.Content(g.Text("Body")),
		Card.Footer(g.Text("Footer")),
	))

	checks := []string{
		`<div class="w-full rounded-lg border bg-card text-card-foreground shadow-xs">`,
		`<h3 class="text-lg font-semibold leading-none tracking-tight">Users</h3>`,
		`<p class="text-sm text-muted-foreground">Manage user accounts.</p>`,
		`<div class="p-6">Body</div>`,
		`<div class="flex items-center p-6 pt-0">Footer</div>`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Card output missing %q in %s", check, got)
		}
	}
}

func TestTableRendersWrapperAndSelectedRow(t *testing.T) {
	got := testutil.RenderNode(t, Table.Root(
		Table.Caption(g.Text("Team members")),
		Table.Header(Table.Row(RowProps{}, Table.Head(g.Text("Name")))),
		Table.Body(BodyProps{ID: "people-table"}, Table.Row(RowProps{Selected: true}, Table.Cell(g.Text("Ada")))),
	))

	checks := []string{
		`<div class="relative w-full overflow-auto"><table class="w-full caption-bottom text-sm">`,
		`<caption class="mt-4 text-sm text-muted-foreground">Team members</caption>`,
		`id="people-table"`,
		`<tbody class="[&amp;_tr:last-child]:border-0"`,
		`<tr class="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted bg-muted">`,
		`<td class="p-2 align-middle`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Table output missing %q in %s", check, got)
		}
	}
}
