package email_test

import (
	"strings"
	"testing"

	"github.com/go-sum/componentry/email"
)

func TestPlainText_CRLFLineEndings(t *testing.T) {
	result := email.PlainText("Line 1", "Line 2", "Line 3")
	if !strings.Contains(result, "\r\n") {
		t.Errorf("expected CRLF line endings, got: %q", result)
	}
}

func TestPlainText_singleLine(t *testing.T) {
	result := email.PlainText("Hello")
	if result != "Hello" {
		t.Errorf("expected %q, got %q", "Hello", result)
	}
}

func TestPlainText_emptyLines(t *testing.T) {
	result := email.PlainText("A", "", "B")
	expected := "A\r\n\r\nB"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
