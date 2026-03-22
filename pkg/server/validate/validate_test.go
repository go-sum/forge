package validate

import "testing"

type validateInput struct {
	Email string `validate:"required,email"`
}

func TestValidatorStructAndVar(t *testing.T) {
	v := New()

	if err := v.Struct(validateInput{Email: "ada@example.com"}); err != nil {
		t.Fatalf("Struct(valid) error = %v", err)
	}
	if err := v.Struct(validateInput{}); err == nil {
		t.Fatal("Struct(invalid) unexpectedly succeeded")
	}

	if err := v.Var("ada@example.com", "required,email"); err != nil {
		t.Fatalf("Var(valid) error = %v", err)
	}
	if err := v.Var("not-an-email", "required,email"); err == nil {
		t.Fatal("Var(invalid) unexpectedly succeeded")
	}

	if v.Validate() == nil {
		t.Fatal("Validate() returned nil")
	}
}
