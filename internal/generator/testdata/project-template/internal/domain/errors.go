package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInactiveUser       = errors.New("user is inactive")
	ErrMissingCurrentUser = errors.New("missing current user")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidInput       = errors.New("invalid input")
)
