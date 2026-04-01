package model

import (
	"errors"

	authmodel "github.com/go-sum/auth/model"
)

var (
	ErrEmailTaken               = authmodel.ErrEmailTaken
	ErrForbidden                = errors.New("forbidden")
	ErrInvalidCredentials       = authmodel.ErrInvalidCredentials
	ErrRequiredRelationsMissing = errors.New("required relations missing")
	ErrUserNotFound             = authmodel.ErrUserNotFound
)
