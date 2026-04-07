package model

import "errors"

var (
	ErrForbidden                = errors.New("forbidden")
	ErrRequiredRelationsMissing = errors.New("required relations missing")
)
