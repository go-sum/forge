// Package form provides request binding and validation for HTML form submissions.
package form

// Form describes the contract for a form submission handler.
// Submission implements this interface.
type Form interface {
	// Submit binds the request body into dest and runs validation.
	IsSubmitted() bool
	IsValid() bool
	// IsDone reports true when the form was submitted and passed validation.
	IsDone() bool
	FieldHasErrors(field string) bool
	GetFieldErrors(field string) []string
	SetFieldError(field, msg string)
	GetErrors() map[string][]string
}
