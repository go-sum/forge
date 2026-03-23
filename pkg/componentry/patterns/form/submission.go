package form

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// Binder decodes a request body into dest.
// echo.Context satisfies this interface via its Bind method, so Echo handlers
// can pass c directly with no adapter required.
type Binder interface {
	Bind(dest any) error
}

// Submission handles a single form POST: binding, validation, and error tracking.
// Construct once per request via New and then call Submit.
type Submission struct {
	v         *validator.Validate
	submitted bool
	errors    map[string][]string
}

const formErrorKey = "_"

// New creates a Submission backed by the provided validator instance.
func New(v *validator.Validate) *Submission {
	return &Submission{
		v:      v,
		errors: make(map[string][]string),
	}
}

// Submit binds the request body into dest via b and validates it.
// Validation errors are stored per-field; binding errors are stored under "_".
// Errors are never propagated out — callers check IsValid() after Submit returns.
func (s *Submission) Submit(b Binder, dest any) {
	s.submitted = true
	if err := b.Bind(dest); err != nil {
		s.SetFormError(err.Error())
		return
	}
	if err := s.v.Struct(dest); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			for _, fe := range verrs {
				field := fe.Field()
				s.errors[field] = append(s.errors[field], fe.Error())
			}
			return
		}
		s.SetFormError(err.Error())
	}
}

func (s *Submission) IsSubmitted() bool { return s.submitted }

func (s *Submission) IsValid() bool { return s.submitted && len(s.errors) == 0 }

func (s *Submission) FieldHasErrors(field string) bool {
	return len(s.errors[field]) > 0
}

func (s *Submission) GetFieldErrors(field string) []string {
	return s.errors[field]
}

func (s *Submission) SetFieldError(field, msg string) {
	s.errors[field] = append(s.errors[field], msg)
}

func (s *Submission) HasFormErrors() bool {
	return len(s.errors[formErrorKey]) > 0
}

func (s *Submission) GetFormErrors() []string {
	return s.errors[formErrorKey]
}

func (s *Submission) SetFormError(msg string) {
	s.SetFieldError(formErrorKey, msg)
}

func (s *Submission) GetErrors() map[string][]string {
	return s.errors
}
