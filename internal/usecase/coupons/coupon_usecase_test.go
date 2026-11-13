package couponuseCase

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

type couponFixture struct {
	uc              *CouponUseCase
	couponRepo      *mocks.MockCouponRepository
	orderRepo       *mocks.MockOrderRepository
	couponUsageRepo *mocks.MockCouponUsageRepository
}

func newCouponFixture(t *testing.T) *couponFixture {
	t.Helper()
	couponRepo := mocks.NewMockCouponRepository(t)
	orderRepo := mocks.NewMockOrderRepository(t)
	couponUsageRepo := mocks.NewMockCouponUsageRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &couponFixture{
		uc:              NewCouponUseCase(couponRepo, orderRepo, couponUsageRepo, log),
		couponRepo:      couponRepo,
		orderRepo:       orderRepo,
		couponUsageRepo: couponUsageRepo,
	}
}

func TestCouponUseCase_CRUDAndToggle(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateCoupon validation duplicate and success", func(t *testing.T) {
		f := newCouponFixture(t)

		nilRes, err := f.uc.CreateCoupon(ctx, nil)
		assert.Nil(t, nilRes)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidCoupon)

		emptyCode := newCoupon()
		emptyCode.Code = "   "
		_, err = f.uc.CreateCoupon(ctx, emptyCode)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidCouponCode)

		dup := newCoupon()
		f.couponRepo.EXPECT().GetByCode(ctx, "HELLO10").Return(&models.Coupon{ID: uuid.New()}, nil).Once()
		_, err = f.uc.CreateCoupon(ctx, dup)
		assert.ErrorIs(t, err, domainerrors.ErrCouponAlreadyExists)

		valid := newCoupon()
		valid.Code = " hello10 "
		f.couponRepo.EXPECT().GetByCode(ctx, "HELLO10").Return(nil, domainerrors.ErrRecordNotFound).Once()
		f.couponRepo.EXPECT().Create(ctx, mock.MatchedBy(func(c *models.Coupon) bool {
			return c.Code == "HELLO10"
		})).Return(nil).Once()

		got, err := f.uc.CreateCoupon(ctx, valid)
		require.NoError(t, err)
		assert.Equal(t, "HELLO10", got.Code)
	})

	t.Run("GetCoupon UpdateCoupon DeleteCoupon ToggleCouponActive ListCoupons", func(t *testing.T) {
		f := newCouponFixture(t)
		_, err := f.uc.GetCoupon(ctx, uuid.Nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidCouponID)

		c := newCoupon()
		f.couponRepo.EXPECT().GetByID(ctx, c.ID).Return(c, nil).Once()
		got, err := f.uc.GetCoupon(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, c.ID, got.ID)

		assert.ErrorIs(t, f.uc.DeleteCoupon(ctx, uuid.Nil), domainerrors.ErrInvalidCouponID)
		f.couponRepo.EXPECT().Delete(ctx, c.ID).Return(nil).Once()
		require.NoError(t, f.uc.DeleteCoupon(ctx, c.ID))

		assert.ErrorIs(t, f.uc.ToggleCouponActive(ctx, uuid.Nil), domainerrors.ErrInvalidCouponID)
		c.IsActive = true
		f.couponRepo.EXPECT().GetByID(ctx, c.ID).Return(c, nil).Once()
		f.couponRepo.EXPECT().SetActive(ctx, c.ID, false).Return(nil).Once()
		require.NoError(t, f.uc.ToggleCouponActive(ctx, c.ID))

		updatedInput := newCoupon()
		updatedInput.ID = c.ID
		updatedInput.Code = "SAVE20"
		updatedInput.DiscountValue = 20
		f.couponRepo.EXPECT().GetByID(ctx, c.ID).Return(c, nil).Once()
		f.couponRepo.EXPECT().GetByCode(ctx, "SAVE20").Return(nil, domainerrors.ErrRecordNotFound).Once()
		f.couponRepo.EXPECT().Update(ctx, mock.MatchedBy(func(cp *models.Coupon) bool {
			return cp.Code == "SAVE20" && cp.DiscountValue == 20
		})).Return(nil).Once()
		updated, err := f.uc.UpdateCoupon(ctx, updatedInput)
		require.NoError(t, err)
		assert.Equal(t, "SAVE20", updated.Code)

		f.couponRepo.EXPECT().List(ctx, "SAVE", (*bool)(nil), 10, 1, (*string)(nil)).
			Return([]models.Coupon{*c}, int64(1), nil).Once()
		list, total, err := f.uc.ListCoupons(ctx, " SAVE ", nil, 0, 0, nil)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, int64(1), total)
	})
}

func TestCouponUseCase_ValidateRedeemAndUsages(t *testing.T) {
	ctx := context.Background()

	t.Run("ValidateCoupon", func(t *testing.T) {
		f := newCouponFixture(t)
		_, _, err := f.uc.ValidateCoupon(ctx, "  ", 1, 500)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, _, err = f.uc.ValidateCoupon(ctx, "HELLO10", 0, 500)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		c := newCoupon()
		f.couponRepo.EXPECT().GetByCode(ctx, "HELLO10").Return(c, nil).Once()
		f.couponUsageRepo.EXPECT().CountByCouponID(ctx, c.ID).Return(int64(1), nil).Once()
		f.couponUsageRepo.EXPECT().CountByCouponAndUser(ctx, c.ID, uint(7)).Return(int64(0), nil).Once()

		gotCoupon, discount, err := f.uc.ValidateCoupon(ctx, " hello10 ", 7, 1000)
		require.NoError(t, err)
		assert.Equal(t, c.ID, gotCoupon.ID)
		assert.Equal(t, 100.0, discount)
	})

	t.Run("RedeemCoupon", func(t *testing.T) {
		f := newCouponFixture(t)
		c := newCoupon()
		orderID := uuid.New()

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 7, Subtotal: 1000, TotalAmount: 900}, nil).Once()
		f.couponRepo.EXPECT().GetByID(ctx, uuid.Nil).Return(nil, domainerrors.ErrRecordNotFound).Once()
		assert.ErrorIs(t, f.uc.RedeemCoupon(ctx, uuid.Nil, 7, orderID), domainerrors.ErrRecordNotFound)

		f.orderRepo.EXPECT().GetByID(ctx, uuid.Nil).Return(nil, domainerrors.ErrRecordNotFound).Once()
		assert.ErrorIs(t, f.uc.RedeemCoupon(ctx, c.ID, 7, uuid.Nil), domainerrors.ErrRecordNotFound)

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{
			ID:          orderID,
			UserID:      7,
			Subtotal:    1000,
			TotalAmount: 900,
		}, nil).Once()
		f.couponRepo.EXPECT().GetByID(ctx, c.ID).Return(c, nil).Once()
		assert.Error(t, f.uc.RedeemCoupon(ctx, c.ID, 0, orderID))

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{
			ID:          orderID,
			UserID:      7,
			Subtotal:    1000,
			TotalAmount: 900,
		}, nil).Once()
		f.couponRepo.EXPECT().GetByID(ctx, c.ID).Return(c, nil).Once()
		f.couponRepo.EXPECT().GetByCode(ctx, "HELLO10").Return(c, nil).Once()
		f.couponUsageRepo.EXPECT().CountByCouponID(ctx, c.ID).Return(int64(0), nil).Once()
		f.couponUsageRepo.EXPECT().CountByCouponAndUser(ctx, c.ID, uint(7)).Return(int64(0), nil).Once()
		f.couponUsageRepo.EXPECT().Create(ctx, mock.MatchedBy(func(u *models.CouponUsage) bool {
			return u.CouponID == c.ID && u.UserID == 7 && u.OrderID == orderID
		})).Return(nil).Once()

		require.NoError(t, f.uc.RedeemCoupon(ctx, c.ID, 7, orderID))
	})

	t.Run("ListCouponUsagesByCouponID and ListCouponUsagesByUserID", func(t *testing.T) {
		f := newCouponFixture(t)
		cID := uuid.New()

		_, _, err := f.uc.ListCouponUsagesByCouponID(ctx, uuid.Nil, 10, 1)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidCouponID)

		f.couponUsageRepo.EXPECT().ListByCouponID(ctx, cID, 10, 1).Return([]models.CouponUsage{{ID: 1}}, int64(1), nil).Once()
		got, total, err := f.uc.ListCouponUsagesByCouponID(ctx, cID, 0, 0)
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, int64(1), total)

		_, _, err = f.uc.ListCouponUsagesByUserID(ctx, 0, 10, 1)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidUserID)

		f.couponUsageRepo.EXPECT().ListByUserID(ctx, uint(7), 10, 1).Return([]models.CouponUsage{{ID: 2}}, int64(1), nil).Once()
		gotU, totalU, err := f.uc.ListCouponUsagesByUserID(ctx, 7, 0, 0)
		require.NoError(t, err)
		assert.Len(t, gotU, 1)
		assert.Equal(t, int64(1), totalU)
	})
}
