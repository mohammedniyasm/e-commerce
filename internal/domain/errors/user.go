package errors

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")

	ErrInvalidName        = errors.New("invalid name")
	ErrInvalidPhone       = errors.New("invalid phone")
	ErrWeakPassword       = errors.New("weak password")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBlocked        = errors.New("user account is blocked")
	ErrRefreshSessionNotFound  = errors.New("refresh session not found")
)
