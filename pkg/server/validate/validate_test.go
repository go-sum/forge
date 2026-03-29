package validate

import "testing"

// Compile-time check: *Validator satisfies Echo v5's Validator interface
// and form.StructValidator via structural typing.
var _ interface{ Validate(any) error } = (*Validator)(nil)

type validateInput struct {
	Email string `validate:"required,email"`
}

func TestValidate(t *testing.T) {
	v := New()

	if err := v.Validate(validateInput{Email: "ada@example.com"}); err != nil {
		t.Fatalf("Validate(valid) error = %v", err)
	}
	if err := v.Validate(validateInput{}); err == nil {
		t.Fatal("Validate(invalid: missing required) unexpectedly succeeded")
	}
	if err := v.Validate(validateInput{Email: "not-an-email"}); err == nil {
		t.Fatal("Validate(invalid: bad email) unexpectedly succeeded")
	}
}

func TestVar(t *testing.T) {
	v := New()

	if err := v.Var("ada@example.com", "required,email"); err != nil {
		t.Fatalf("Var(valid) error = %v", err)
	}
	if err := v.Var("not-an-email", "required,email"); err == nil {
		t.Fatal("Var(invalid) unexpectedly succeeded")
	}
}
