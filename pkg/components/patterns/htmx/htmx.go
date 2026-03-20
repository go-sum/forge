// Package htmx provides typed attribute builders plus request/response helpers
// for server-first HTMX components. It depends only on net/http and pkg/components
// tiers below patterns.
package htmx

import "net/http"

// Request inspection helpers.

func IsRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func IsBoosted(r *http.Request) bool {
	return r.Header.Get("HX-Boosted") == "true"
}

func GetTrigger(r *http.Request) string {
	return r.Header.Get("HX-Trigger")
}

func GetTarget(r *http.Request) string {
	return r.Header.Get("HX-Target")
}

func GetTriggerName(r *http.Request) string {
	return r.Header.Get("HX-Trigger-Name")
}

func GetCurrentURL(r *http.Request) string {
	return r.Header.Get("HX-Current-URL")
}

// Response header helpers.

func SetRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", url)
}

func SetRefresh(w http.ResponseWriter) {
	w.Header().Set("HX-Refresh", "true")
}

func SetPushURL(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Push-Url", url)
}

func SetReplaceURL(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Replace-Url", url)
}

func SetTrigger(w http.ResponseWriter, event string) {
	w.Header().Set("HX-Trigger", event)
}

func SetTriggerAfterSettle(w http.ResponseWriter, event string) {
	w.Header().Set("HX-Trigger-After-Settle", event)
}

func SetRetarget(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Retarget", selector)
}

func SetReswap(w http.ResponseWriter, strategy string) {
	w.Header().Set("HX-Reswap", strategy)
}
