package service

import "errors"

var (
	ErrNotFound        = errors.New("record not found")
	ErrAppNameExists   = errors.New("application name already exists")
	ErrIfaceNameExists = errors.New("interface name already exists")
	ErrInvalidOptions  = errors.New("options validation failed")
	ErrValidation      = errors.New("validation failed")
)
