package errors

import "errors"

var (
	ErrRecordNotFound = errors.New("Record not found")

	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")

	ErrInvalidProfileImage = errors.New("invalid profile image")
	ErrInvalidName         = errors.New("invalid name")
	ErrInvalidPhone        = errors.New("invalid phone")
	ErrWeakPassword        = errors.New("weak password")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidOTP          = errors.New("invalid otp")
	ErrInvalidAddress      = errors.New("invalid address")
	ErrInvalidUserID       = errors.New("invalid user id")

	ErrUserBlocked            = errors.New("user account is blocked")
	ErrRefreshSessionNotFound = errors.New("refresh session not found")
	ErrEmailAlreadyVerified   = errors.New("email is already verified")

	ErrInvalidCategoryName          = errors.New("invalid category name")
	ErrCategoryAlreadyExists        = errors.New("category already exists")
	ErrCategoryNameAlreadyExists    = errors.New("category name already used")
	ErrInvalidCategoryID            = errors.New("invalid category id")
	ErrRestoreCategoryAlreadyExists = errors.New("Category already exists, can't restore category")

	ErrInvalidBrandName          = errors.New("invalid brand name")
	ErrBrandAlreadyExists        = errors.New("brand already exists")
	ErrBrandNameAlreadyExists    = errors.New("brand name already used")
	ErrInvalidBrandID            = errors.New("invalid brand id")
	ErrRestoreBrandAlreadyExists = errors.New("brand already exists, can't restore brand")
	ErrInvalidBrandLogo          = errors.New("invalid brand logo")
	ErrInvalidBrandLogoSize      = errors.New("invalid brand logo size")
)
