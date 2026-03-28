package model

import (
	"errors"

	authmodel "github.com/go-sum/auth/model"
)

var (
	ErrUserNotFound       = authmodel.ErrUserNotFound
	ErrEmailTaken         = authmodel.ErrEmailTaken
	ErrInvalidCredentials = authmodel.ErrInvalidCredentials
	ErrForbidden          = errors.New("forbidden")
)
