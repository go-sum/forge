// Package validate defines the validator port that auth requires from the server module.
package validate

import "github.com/go-playground/validator/v10"

// Validator is the narrow interface auth requires from the validate package.
// *server/validate.Validator satisfies this interface (confirmed in foundry's adapters).
type Validator interface {
	Validate() *validator.Validate
}
