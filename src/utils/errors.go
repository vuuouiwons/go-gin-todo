package utils

import "errors"

var (
	// ErrNotFound indicates the requested resource does not exist or user lacks access.
	ErrNotFound = errors.New("resource not found")

	// ErrConflict indicates a unique constraint violation or duplicate resource.
	ErrConflict = errors.New("resource already exists")

	// ErrBadRequest indicates invalid input, missing parameters, or validation failure.
	ErrBadRequest = errors.New("invalid request parameters")

	// ErrUnauthorized indicates missing, expired, or invalid authentication credentials.
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrInternalServer indicates an unexpected infrastructure or persistence error.
	ErrInternalServer = errors.New("internal server error")
)
