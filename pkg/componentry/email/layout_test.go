package email_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-sum/componentry/email"

	g "maragu.dev/gomponents"
)

func TestLayout_containsTitle(t *testing.T) {
	node := email.Layout("Test Email", email.P("Hello"))
	var buf bytes.Buffer
	if err := node.Render(&buf); err != nil {
		t.Fatalf("Render returned unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "<title>Test Email</title>") {
		t.Errorf("expected <title>Test Email</title> in output, got:\n%s", out)
	}
}

func TestLayout_containsBody(t *testing.T) {
	node := email.Layout("Subject", email.P("My content"))
	var buf bytes.Buffer
	_ = node.Render(&buf)
	if !strings.Contains(buf.String(), "My content") {
		t.Error("expected body content in output")
	}
}

func TestLayout_tableStructure(t *testing.T) {
	node := email.Layout("T", g.Text(""))
	var buf bytes.Buffer
	_ = node.Render(&buf)
	out := buf.String()
	if !strings.Contains(out, "<table") {
		t.Error("expected <table element in email layout output")
	}
}

func TestH1(t *testing.T) {
	node := email.H1("Welcome!")
	var buf bytes.Buffer
	_ = node.Render(&buf)
	if !strings.Contains(buf.String(), "Welcome!") {
		t.Error("expected heading text in output")
	}
}

func TestP(t *testing.T) {
	node := email.P("A paragraph.")
	var buf bytes.Buffer
	_ = node.Render(&buf)
	out := buf.String()
	if !strings.Contains(out, "<p") || !strings.Contains(out, "A paragraph.") {
		t.Errorf("expected <p> with text in output, got: %s", out)
	}
}
