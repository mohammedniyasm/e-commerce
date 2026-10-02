package errors

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidProfileImage = errors.New("invalid profile image")