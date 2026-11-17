package domain

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrConflict       = errors.New("resource already exists")
	ErrBadRequest     = errors.New("invalid input data")
	ErrInternalServer = errors.New("internal server error")
)
