package model

import "errors"

var (
	ErrAdminExists              = errors.New("admin already exists")
	ErrForbidden                = errors.New("forbidden")
	ErrRequiredRelationsMissing = errors.New("required relations missing")
)
