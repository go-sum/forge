package page

import (
	"strings"
	"testing"

	"starter/internal/model"
	"starter/internal/view"
	pkgform "starter/pkg/components/patterns/form"
	"starter/pkg/components/testutil"

	"github.com/go-playground/validator/v10"
)

func TestLoginPageRendersInputAndErrors(t *testing.T) {
	form := pkgform.New(validator.New())
	form.SetFieldError("Email", "Email is required.")
	form.SetFormError("Invalid email or password.")

	got := testutil.RenderNode(t, LoginPage(view.Request{
		CSRFToken: "csrf-token",
	}, form, model.LoginInput{Email: "ada@example.com"}))

	wantSnippets := []string{
		`action="/login"`,
		`value="ada@example.com"`,
		`value="csrf-token"`,
		`Email is required.`,
		`Invalid email or password.`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered login page missing %q:\n%s", want, got)
		}
	}
}

func TestRegisterPageRendersInputAndErrors(t *testing.T) {
	form := pkgform.New(validator.New())
	form.SetFieldError("DisplayName", "Display name is required.")
	form.SetFormError("Unable to save account.")

	got := testutil.RenderNode(t, RegisterPage(view.Request{
		CSRFToken: "csrf-token",
	}, form, model.CreateUserInput{
		Email:       "ada@example.com",
		DisplayName: "Ada",
	}))

	wantSnippets := []string{
		`action="/register"`,
		`value="ada@example.com"`,
		`value="Ada"`,
		`Display name is required.`,
		`Unable to save account.`,
	}
	for _, want := range wantSnippets {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered register page missing %q:\n%s", want, got)
		}
	}
}
