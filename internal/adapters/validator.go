// Package adapters contains translation code connecting forge's concrete types
// to the interface ports defined by the auth and server packages.
package adapters

import (
	authvalidate "github.com/go-sum/auth/validate"
	"github.com/go-sum/server/validate"
)

// Compile-time assertion: *validate.Validator satisfies auth's Validator port.
var _ authvalidate.Validator = (*validate.Validator)(nil)
