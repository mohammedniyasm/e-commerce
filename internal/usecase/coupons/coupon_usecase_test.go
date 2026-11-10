package couponuseCase

import (
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newCoupon() *models.Coupon {
	max := 500.00
	lim := 100
	return &models.Coupon{
		ID:                 uuid.New(),
		Code:               "Hello10",
		DiscountType:       "percentage",
		MinimumOrderAmount: 100,
		MaximumDiscount:    &max,
		DiscountValue:      10,
		UsageLimit:         &lim,
		UsageLimitPerUser: func() *int {
			v := 2
			return &v
		}(),
		StartAt: func() time.Time {
			ti := time.Now().Add(-72 * time.Hour)
			return ti
		}(),
		ExpiresAt: func() time.Time {
			ti := time.Now().Add(72 * time.Hour)
			return ti
		}(),
		IsActive: true,
	}
}

type testCase struct {
	name        string
	coupon      *models.Coupon
	orderAmount float64
	discount    float64
	wantError   error
}

func TestValidateCoupon(t *testing.T) {
	tests := []testCase{
		testCase{
			name:      "valid coupon",
			coupon:    newCoupon(),
			wantError: nil,
		},
		testCase{
			name:      "invalid coupon",
			coupon:    nil,
			wantError: domainerrors.ErrInvalidCoupon,
		},
		testCase{
			name: "invalid coupon code",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.Code = "    "
				return c
			}(),
			wantError: domainerrors.ErrInvalidCouponCode,
		},
		testCase{
			name: "invalid coupon type",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.DiscountType = "something"
				return c
			}(),
			wantError: domainerrors.ErrCouponDiscountInvalid,
		},
		testCase{
			name: "invalid coupon value in percentage (exceed)",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.DiscountValue = 101
				return c
			}(),
			wantError: domainerrors.ErrCouponDiscountInvalid,
		},
		testCase{
			name: "invalid coupon value in percentage",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.DiscountValue = 0
				return c
			}(),
			wantError: domainerrors.ErrCouponDiscountInvalid,
		},
		testCase{
			name: "invalid coupon value in fixed",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.DiscountType = "fixed"
				c.DiscountValue = 0
				return c
			}(),
			wantError: domainerrors.ErrCouponDiscountInvalid,
		},
		testCase{
			name: "invalid minimum order amount",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.MinimumOrderAmount = -1
				return c
			}(),
			wantError: domainerrors.ErrInvalidCoupon,
		},
		testCase{
			name: "invalid maximum order amount",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.MinimumOrderAmount = -1
				return c
			}(),
			wantError: domainerrors.ErrInvalidCoupon,
		},
		testCase{
			name: "invalid usage limit",
			coupon: func() *models.Coupon {
				c := newCoupon()
				*c.UsageLimit = -1
				return c
			}(),
			wantError: domainerrors.ErrInvalidCoupon,
		},
		testCase{
			name: "invalid usage limit per user",
			coupon: func() *models.Coupon {
				c := newCoupon()
				*c.UsageLimit = -1
				return c
			}(),
			wantError: domainerrors.ErrInvalidCoupon,
		},
		testCase{
			name: "invalid start time ",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.StartAt = time.Now().Add(100 * time.Hour)
				return c
			}(),
			wantError: domainerrors.ErrInvalidCoupon,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCoupon(tt.coupon)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("expected: %v,got: %v", err, tt.wantError)
			}
		})
	}
}
func TestCalculateDiscount(t *testing.T) {
	tests := []testCase{
		{
			name:        "valid percentage coupon",
			coupon:      newCoupon(),
			orderAmount: 500,
			discount:    50,
			wantError:   nil,
		},
		{
			name: "valid fixed coupon",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.DiscountType = "fixed"
				c.DiscountValue = 100
				return c
			}(),
			orderAmount: 500,
			discount:    100,
			wantError:   nil,
		},
		{
			name: "percentage discount exceeds maximum discount",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.DiscountValue = 20
				max := 500.0
				c.MaximumDiscount = &max
				return c
			}(),
			orderAmount: 10000,
			discount:    500,
			wantError:   nil,
		},
		{
			name: "fixed discount exceeds order amount",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.DiscountType = "fixed"
				c.DiscountValue = 200
				return c
			}(),
			orderAmount: 100,
			discount:    100,
			wantError:   nil,
		},
		{
			name: "maximum discount is nil",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.MaximumDiscount = nil
				return c
			}(),
			orderAmount: 1000,
			discount:    100,
			wantError:   nil,
		},
		{
			name: "discount equals maximum discount",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.DiscountValue = 100
				max := 500.0
				c.MaximumDiscount = &max
				return c
			}(),
			orderAmount: 500,
			discount:    500,
			wantError:   nil,
		},
		{
			name: "invalid discount type",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.DiscountType = "something"
				return c
			}(),
			orderAmount: 500,
			discount:    0,
			wantError:   domainerrors.ErrCouponDiscountInvalid,
		},
		{
			name: "zero discount value",
			coupon: func() *models.Coupon {
				c := newCoupon()
				c.DiscountValue = 0
				return c
			}(),
			orderAmount: 500,
			discount:    0,
			wantError:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discount, err := calculateDiscount(tt.coupon, tt.orderAmount)
			if discount != tt.discount {
				t.Errorf("expected discount: %v,got: %v", tt.discount, discount)
			}
			if !errors.Is(err, tt.wantError) {
				t.Errorf("expected: %v,got: %v", err, tt.wantError)
			}
		})
	}
}
