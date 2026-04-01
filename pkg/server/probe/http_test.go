package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPGet(t *testing.T) {
	t.Run("passes on 2xx", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		if err := HTTPGet(context.Background(), srv.Client(), srv.URL); err != nil {
			t.Fatalf("HTTPGet() error = %v", err)
		}
	})

	t.Run("fails on non success status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer srv.Close()

		err := HTTPGet(context.Background(), srv.Client(), srv.URL)
		if err == nil {
			t.Fatal("HTTPGet() error = nil")
		}
		want := "GET " + srv.URL + ": status 503"
		if err.Error() != want {
			t.Fatalf("HTTPGet() error = %q, want %q", err.Error(), want)
		}
	})

	t.Run("fails on transport error", func(t *testing.T) {
		err := HTTPGet(context.Background(), http.DefaultClient, "http://127.0.0.1:1")
		if err == nil {
			t.Fatal("HTTPGet() error = nil")
		}
		if !strings.Contains(err.Error(), "GET http://127.0.0.1:1") {
			t.Fatalf("HTTPGet() error = %q, want GET prefix", err.Error())
		}
	})
}
