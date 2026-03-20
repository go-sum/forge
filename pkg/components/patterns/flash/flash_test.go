package flash

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"starter/pkg/components/testutil"
)

func TestSetAndGetAllRoundTripMessages(t *testing.T) {
	setRecorder := httptest.NewRecorder()
	msgs := []Message{{Type: TypeSuccess, Text: "Saved"}}
	if err := Set(setRecorder, msgs); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	resp := setRecorder.Result()
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Set() cookies = %d, want 1", len(cookies))
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookies[0])
	getRecorder := httptest.NewRecorder()
	got, err := GetAll(req, getRecorder)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(got) != 1 || got[0].Type != TypeSuccess || got[0].Text != "Saved" {
		t.Fatalf("GetAll() = %#v", got)
	}
	if cleared := getRecorder.Result().Cookies(); len(cleared) != 1 || cleared[0].MaxAge != -1 {
		t.Fatalf("GetAll() did not clear cookie: %#v", cleared)
	}
}

func TestRenderOOBAppendsToToastContainer(t *testing.T) {
	got := testutil.RenderNode(t, RenderOOB([]Message{{Type: TypeSuccess, Text: "Saved"}}))
	if !strings.Contains(got, `hx-swap-oob="beforeend:#toast-container"`) || !strings.Contains(got, `Saved`) {
		t.Fatalf("RenderOOB() output missing out-of-band toast markup in %s", got)
	}
}
