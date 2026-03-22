package htmx

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-sum/componentry/testutil"
	"github.com/go-sum/componentry/ui/feedback"

	g "maragu.dev/gomponents"
)

func TestRequestAndResponseHelpers(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Boosted", "true")
	req.Header.Set("HX-Trigger", "save")
	req.Header.Set("HX-Target", "#row-1")
	req.Header.Set("HX-Trigger-Name", "submit")
	req.Header.Set("HX-Current-URL", "https://example.test/users")
	rec := httptest.NewRecorder()

	htmxReq := NewRequest(req)
	if !htmxReq.Enabled || !htmxReq.Boosted || htmxReq.Target != "#row-1" {
		t.Fatalf("NewRequest() = %#v", htmxReq)
	}
	if htmxReq.IsPartial() {
		t.Fatalf("boosted request unexpectedly treated as partial: %#v", htmxReq)
	}

	if !IsRequest(req) || !IsBoosted(req) {
		t.Fatal("request helpers did not detect HTMX headers")
	}
	if GetTrigger(req) != "save" || GetTarget(req) != "#row-1" || GetTriggerName(req) != "submit" || GetCurrentURL(req) != "https://example.test/users" {
		t.Fatal("request helper values were not preserved")
	}

	SetRedirect(rec, "/login")
	SetRefresh(rec)
	SetPushURL(rec, "/users?page=2")
	SetReplaceURL(rec, "/users?page=3")
	SetTrigger(rec, "users:updated")
	SetTriggerAfterSettle(rec, "users:settled")
	SetTriggerAfterSwap(rec, "users:swapped")
	SetRetarget(rec, "#users")
	SetReswap(rec, "outerHTML")

	headers := rec.Header()
	checks := map[string]string{
		"HX-Redirect":             "/login",
		"HX-Refresh":              "true",
		"HX-Push-Url":             "/users?page=2",
		"HX-Replace-Url":          "/users?page=3",
		"HX-Trigger":              "users:updated",
		"HX-Trigger-After-Settle": "users:settled",
		"HX-Trigger-After-Swap":   "users:swapped",
		"HX-Retarget":             "#users",
		"HX-Reswap":               "outerHTML",
	}
	for key, want := range checks {
		if got := headers.Get(key); got != want {
			t.Fatalf("header %s = %q, want %q", key, got, want)
		}
	}
}

func TestAttrsAndPatternsRenderTypedHTMXMarkup(t *testing.T) {
	boost := true
	got := testutil.RenderNode(t, g.El("div",
		Attrs(AttrsProps{
			Get:         "/users",
			Target:      "#users-table",
			Swap:        SwapOuterHTML,
			Trigger:     "load",
			Values:      map[string]string{"page": "2"},
			Headers:     map[string]string{"X-Test": "1"},
			Boost:       &boost,
			DisabledElt: "this",
		})...,
	))

	checks := []string{
		`hx-get="/users"`,
		`hx-target="#users-table"`,
		`hx-swap="outerHTML"`,
		`hx-trigger="load"`,
		`hx-vals="{&#34;page&#34;:&#34;2&#34;}"`,
		`hx-headers="{&#34;X-Test&#34;:&#34;1&#34;}"`,
		`hx-boost="true"`,
		`hx-disabled-elt="this"`,
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("Attrs() output missing %q in %s", check, got)
		}
	}

	liveSearch := testutil.RenderNode(t, g.El("input", LiveSearch(LiveSearchProps{
		Path:    "/users/search",
		Target:  "#results",
		PushURL: true,
	})...))
	if !strings.Contains(liveSearch, `hx-trigger="input changed delay:300ms, search"`) || !strings.Contains(liveSearch, `hx-push-url="true"`) {
		t.Fatalf("LiveSearch() output missing defaults in %s", liveSearch)
	}

	inlineValidation := testutil.RenderNode(t, g.El("input", InlineValidation(InlineValidationProps{
		Path:   "/users/validate/email",
		Target: "#email-field",
	})...))
	if !strings.Contains(inlineValidation, `hx-sync="closest form:abort"`) || !strings.Contains(inlineValidation, `hx-swap="outerHTML"`) {
		t.Fatalf("InlineValidation() output missing defaults in %s", inlineValidation)
	}

	pagination := testutil.RenderNode(t, g.El("a", PaginatedTableLink(PaginatedTableProps{
		Path:      "/users?role=admin",
		Page:      3,
		Target:    "#users-table",
		PushURL:   true,
		Indicator: "#spinner",
	})...))
	if !strings.Contains(pagination, `hx-get="/users?page=3&amp;role=admin"`) || !strings.Contains(pagination, `hx-indicator="#spinner"`) {
		t.Fatalf("PaginatedTableLink() output missing query/indicator in %s", pagination)
	}

	dialogTrigger := testutil.RenderNode(t, g.El("button", AsyncDialogTrigger(AsyncDialogProps{
		Path:     "/users/new",
		DialogID: "user-dialog",
		Target:   "#user-dialog-body",
	})...))
	if !strings.Contains(dialogTrigger, `data-dialog-open="user-dialog"`) || !strings.Contains(dialogTrigger, `hx-get="/users/new"`) {
		t.Fatalf("AsyncDialogTrigger() output missing dialog attrs in %s", dialogTrigger)
	}

	dependentSelect := testutil.RenderNode(t, g.El("select", DependentSelect(DependentSelectProps{
		Path:   "/roles/options",
		Target: "#role-options",
	})...))
	if !strings.Contains(dependentSelect, `hx-trigger="change"`) || !strings.Contains(dependentSelect, `hx-target="#role-options"`) {
		t.Fatalf("DependentSelect() output missing defaults in %s", dependentSelect)
	}
}

func TestOOBHelpersRenderToastMarkup(t *testing.T) {
	swap := testutil.RenderNode(t, g.El("div", OOBAppend("#toast-container")...))
	if !strings.Contains(swap, `hx-swap-oob="beforeend:#toast-container"`) {
		t.Fatalf("OOBAppend() output missing swap attribute in %s", swap)
	}

	toast := testutil.RenderNode(t, ToastOOB(ToastOOBProps{
		Toast: feedback.ToastProps{
			Description: "Saved",
			Variant:     feedback.ToastSuccess,
			Dismissible: true,
		},
	}))
	if !strings.Contains(toast, `hx-swap-oob="beforeend:#toast-container"`) || !strings.Contains(toast, `Saved`) {
		t.Fatalf("ToastOOB() output missing out-of-band toast markup in %s", toast)
	}
}
