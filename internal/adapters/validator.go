// Package adapters contains translation code connecting foundry's concrete types
// to the interface ports defined by the auth and server packages.
package adapters

import (
	authvalidate "github.com/y-goweb/auth/validate"
	"github.com/y-goweb/server/validate"
)

// Compile-time assertion: *validate.Validator satisfies auth's Validator port.
var _ authvalidate.Validator = (*validate.Validator)(nil)
