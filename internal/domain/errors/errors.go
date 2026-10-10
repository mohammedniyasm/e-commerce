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

	ErrInvalidProductID   = errors.New("invalid product id")
	ErrInvalidProductName = errors.New("product name is required")
	ErrInvalidProductSlug = errors.New("product slug is required")
	ErrProductSlugExists  = errors.New("product slug already exists")

	ErrCategoryInactive = errors.New("category is inactive")
	ErrBrandInactive    = errors.New("brand is inactive")

	ErrInvalidVariantID        = errors.New("invalid variant id")
	ErrInvalidVariantSKU       = errors.New("variant SKU is required")
	ErrVariantSKUAlreadyExists = errors.New("variant SKU already exists")
	ErrVariantAlreadyExists    = errors.New("variant with the same size and color already exists")

	ErrInvalidMRP                 = errors.New("MRP must be greater than zero")
	ErrInvalidSellingPrice        = errors.New("selling price must be greater than zero")
	ErrSellingPriceGreaterThanMRP = errors.New("selling price cannot be greater than MRP")
	ErrInvalidStock               = errors.New("stock cannot be negative")

	ErrInvalidProductImageID         = errors.New("invalid product image id")
	ErrProductImagesRequired         = errors.New("product images are required")
	ErrMinimumProductImages          = errors.New("product must have at least 3 images")
	ErrInvalidProductImage           = errors.New("invalid product image")
	ErrInvalidProductImageSize       = errors.New("product image size is invalid")
	ErrInvalidDisplayOrder           = errors.New("invalid display order")
	ErrVariantDoesNotBelongToProduct = errors.New("variant does not belong to product")

	ErrCategoryHasActiveProducts = errors.New("category has active products")
	ErrBrandHasActiveProducts    = errors.New("brand has active products")

	ErrInvalidProductSort  = errors.New("invalid product sort")
	ErrInvalidPriceRange   = errors.New("invalid price range")
	ErrInvalidProductPage  = errors.New("invalid product page")
	ErrInvalidProductLimit = errors.New("invalid product limit")
)
