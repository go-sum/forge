// Package validate wraps go-playground/validator with a reusable Validator type.
package validate

import "github.com/go-playground/validator/v10"

// Validator wraps the underlying go-playground validator instance.
// Callers should construct a single instance at startup and pass it wherever needed.
type Validator struct {
	v *validator.Validate
}

// New creates a ready-to-use Validator.
func New() *Validator {
	return &Validator{v: validator.New()}
}

// Struct validates a struct using its validate tags.
// Returns a validator.ValidationErrors slice on failure.
func (vl *Validator) Struct(s any) error {
	return vl.v.Struct(s)
}

// Var validates a single variable against the given tag expression.
func (vl *Validator) Var(field any, tag string) error {
	return vl.v.Var(field, tag)
}
