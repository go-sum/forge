// Package validate wraps go-playground/validator with a reusable Validator type.
package validate

import "github.com/go-playground/validator/v10"

// Validator wraps the underlying go-playground validator instance.
// Callers should construct a single instance at startup and pass it wherever needed.
//
// *Validator satisfies Echo v5's Validator interface and form.StructValidator
// via its Validate method.
type Validator struct {
	v *validator.Validate
}

// New creates a ready-to-use Validator.
func New() *Validator {
	return &Validator{v: validator.New()}
}

// Validate validates a struct against its validate tags.
// Satisfies Echo v5's Validator interface and form.StructValidator via structural typing.
func (vl *Validator) Validate(i any) error {
	return vl.v.Struct(i)
}

// Var validates a single variable against the given tag expression.
func (vl *Validator) Var(field any, tag string) error {
	return vl.v.Var(field, tag)
}
