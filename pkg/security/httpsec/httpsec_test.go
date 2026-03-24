package httpsec

import (
	"net/http"
	"testing"
)

func TestIsSafeMethod(t *testing.T) {
	if !IsSafeMethod(http.MethodGet) {
		t.Fatal("GET should be safe")
	}
	if IsSafeMethod(http.MethodPost) {
		t.Fatal("POST should not be safe")
	}
}
