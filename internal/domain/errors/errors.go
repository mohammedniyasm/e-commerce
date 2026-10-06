package errors

import "errors"

var (
	ErrEmailAlreadyExists        = errors.New("email already exists")
	ErrCategoryAlreadyExists     = errors.New("category already exists")
	ErrCategoryNameAlreadyExists = errors.New("category name already used")
	ErrUserNotFound              = errors.New("user not found")
	ErrRecordNotFound            = errors.New("Record not found")
	ErrInvalidProfileImage       = errors.New("invalid profile image")
	ErrInvalidName               = errors.New("invalid name")
	ErrInvalidCategoryName       = errors.New("invalid category name")
	ErrInvalidPhone              = errors.New("invalid phone")
	ErrWeakPassword              = errors.New("weak password")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrUserBlocked               = errors.New("user account is blocked")
	ErrRefreshSessionNotFound    = errors.New("refresh session not found")
	ErrInvalidOTP                = errors.New("invalid otp")
	ErrEmailAlreadyVerified      = errors.New("email is already verified")
	ErrInvalidUserID             = errors.New("invalid user id")
	ErrInvalidAddress            = errors.New("invalid address")
	ErrInvalidCategoryID         = errors.New("invalid category id")
	ErrRestoreCategoryAlreadyExists = errors.New("Category already exists, can't restore category")
)
