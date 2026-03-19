package htmx

import (
	"net/http/httptest"
	"testing"
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
		"HX-Retarget":             "#users",
		"HX-Reswap":               "outerHTML",
	}
	for key, want := range checks {
		if got := headers.Get(key); got != want {
			t.Fatalf("header %s = %q, want %q", key, got, want)
		}
	}
}
