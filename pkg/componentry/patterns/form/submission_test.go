package form

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestSubmissionTracksFieldErrors(t *testing.T) {
	s := New(validator.New())
	if s.IsSubmitted() || s.IsValid() {
		t.Fatal("new submission should start empty")
	}

	s.SetFieldError("Email", "required")
	if !s.FieldHasErrors("Email") {
		t.Fatal("SetFieldError() did not mark field")
	}
	if got := s.GetFieldErrors("Email"); len(got) != 1 || got[0] != "required" {
		t.Fatalf("GetFieldErrors() = %#v", got)
	}
	if len(s.GetErrors()) != 1 {
		t.Fatalf("GetErrors() = %#v", s.GetErrors())
	}
}

// stubBinder is a test double for the Binder interface.
type stubBinder struct{ err error }

func (sb *stubBinder) Bind(dest any) error { return sb.err }

func TestSubmitStoresBindError(t *testing.T) {
	s := New(validator.New())
	bindErr := errors.New("malformed body")

	s.Submit(&stubBinder{err: bindErr}, &struct{}{})
	if !s.IsSubmitted() {
		t.Fatal("IsSubmitted() should be true after Submit")
	}
	if s.IsValid() {
		t.Fatal("IsValid() should be false when bind failed")
	}
	if errs := s.GetFormErrors(); len(errs) != 1 || errs[0] != bindErr.Error() {
		t.Fatalf("GetFormErrors() = %#v", errs)
	}
}

func TestSubmitValidatesStruct(t *testing.T) {
	type loginForm struct {
		Email string `validate:"required,email"`
	}

	s := New(validator.New())
	// Bind succeeds but dest is zero-value — validator should catch the required field.
	s.Submit(&stubBinder{}, &loginForm{})
	if s.IsValid() {
		t.Fatal("IsValid() should be false when validation fails")
	}
	if !s.FieldHasErrors("Email") {
		t.Fatalf("expected Email field error; got errors = %#v", s.GetErrors())
	}
}

func TestSubmissionTracksFormErrors(t *testing.T) {
	s := New(validator.New())
	s.SetFormError("save failed")
	if !s.HasFormErrors() {
		t.Fatal("HasFormErrors() should report form-level errors")
	}
	if got := s.GetFormErrors(); len(got) != 1 || got[0] != "save failed" {
		t.Fatalf("GetFormErrors() = %#v", got)
	}
}

func TestSubmitStoresNonValidationValidatorErrors(t *testing.T) {
	s := New(validator.New())
	dest := 123 // validator.Struct on a non-struct returns InvalidValidationError.
	s.Submit(&stubBinder{}, dest)
	if s.IsValid() {
		t.Fatal("IsValid() should be false when validator returns a non-validation error")
	}
	if errs := s.GetFormErrors(); len(errs) != 1 {
		t.Fatalf("GetFormErrors() = %#v", errs)
	}
}
