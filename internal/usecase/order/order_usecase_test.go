package order

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIsValidOrderTransition(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		nextStatus    string
		want          bool
	}{
		// Valid transitions
		{
			name:          "pending to confirmed",
			currentStatus: "pending",
			nextStatus:    "confirmed",
			want:          true,
		},
		{
			name:          "pending to cancelled",
			currentStatus: "pending",
			nextStatus:    "cancelled",
			want:          true,
		},
		{
			name:          "confirmed to processing",
			currentStatus: "confirmed",
			nextStatus:    "processing",
			want:          true,
		},
		{
			name:          "confirmed to cancelled",
			currentStatus: "confirmed",
			nextStatus:    "cancelled",
			want:          true,
		},
		{
			name:          "processing to shipped",
			currentStatus: "processing",
			nextStatus:    "shipped",
			want:          true,
		},
		{
			name:          "processing to cancelled",
			currentStatus: "processing",
			nextStatus:    "cancelled",
			want:          true,
		},
		{
			name:          "shipped to delivered",
			currentStatus: "shipped",
			nextStatus:    "delivered",
			want:          true,
		},
		{
			name:          "shipped to returned",
			currentStatus: "shipped",
			nextStatus:    "returned",
			want:          true,
		},
		{
			name:          "out for delivery to delivered",
			currentStatus: "out_for_delivery",
			nextStatus:    "delivered",
			want:          true,
		},
		{
			name:          "out for delivery to cancelled",
			currentStatus: "out_for_delivery",
			nextStatus:    "cancelled",
			want:          true,
		},
		{
			name:          "delivered to returned",
			currentStatus: "delivered",
			nextStatus:    "returned",
			want:          true,
		},
		{
			name:          "delivered to processing",
			currentStatus: "delivered",
			nextStatus:    "processing",
			want:          true,
		},

		// Invalid transitions
		{
			name:          "pending to shipped",
			currentStatus: "pending",
			nextStatus:    "shipped",
			want:          false,
		},
		{
			name:          "confirmed to delivered",
			currentStatus: "confirmed",
			nextStatus:    "delivered",
			want:          false,
		},
		{
			name:          "processing to pending",
			currentStatus: "processing",
			nextStatus:    "pending",
			want:          false,
		},
		{
			name:          "shipped to cancelled",
			currentStatus: "shipped",
			nextStatus:    "cancelled",
			want:          false,
		},
		{
			name:          "delivered to cancelled",
			currentStatus: "delivered",
			nextStatus:    "cancelled",
			want:          false,
		},

		// Edge cases
		{
			name:          "same status transition",
			currentStatus: "pending",
			nextStatus:    "pending",
			want:          false,
		},
		{
			name:          "unknown current status",
			currentStatus: "unknown",
			nextStatus:    "confirmed",
			want:          false,
		},
		{
			name:          "unknown next status",
			currentStatus: "pending",
			nextStatus:    "unknown",
			want:          false,
		},
		{
			name:          "empty current status",
			currentStatus: "",
			nextStatus:    "confirmed",
			want:          false,
		},
		{
			name:          "empty next status",
			currentStatus: "pending",
			nextStatus:    "",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidOrderTransition(tt.currentStatus, tt.nextStatus)

			if got != tt.want {
				t.Errorf(
					"isValidOrderTransition(%q, %q) = %v; want %v",
					tt.currentStatus,
					tt.nextStatus,
					got,
					tt.want,
				)
			}
		})
	}
}

type orderFixture struct {
	uc                 *OrderUseCase
	orderRepo          *mocks.MockOrderRepository
	uow                *mocks.MockUnitOfWork
	invoiceGen         *mocks.MockInvoiceGenerator
	txOrder            *mocks.MockOrderRepository
	txOrderItem        *mocks.MockOrderItemRepository
	txOrderAddr        *mocks.MockOrderAddressRepository
	txHistory          *mocks.MockOrderStatusHistoryRepository
	txPayment          *mocks.MockPaymentRepository
	txShipment         *mocks.MockShipmentRepository
	txAddress          *mocks.MockAddressRepository
	txCart             *mocks.MockCartRepository
	txVariant          *mocks.MockProductVariantRepository
	txWallet           *mocks.MockWalletRepository
	txCoupon           *mocks.MockCouponRepository
	txCouponUse        *mocks.MockCouponUsageRepository
	txUser             *mocks.MockUserRepository
	txNotification     *mocks.MockNotificationRepository
	txReturnRequest    *mocks.MockReturnRequestRepository
	txReferral         *mocks.MockReferralRepository
	txReferralSettings *mocks.MockReferralSettingsRepository
	txWalletTopUp      *mocks.MockWalletTopUpRepository
}

func newOrderFixture(t *testing.T) *orderFixture {
	t.Helper()
	orderRepo := mocks.NewMockOrderRepository(t)
	uow := mocks.NewMockUnitOfWork(t)
	invoiceGen := mocks.NewMockInvoiceGenerator(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &orderFixture{
		uc:                 NewOrderUseCase(orderRepo, uow, invoiceGen, log),
		orderRepo:          orderRepo,
		uow:                uow,
		invoiceGen:         invoiceGen,
		txOrder:            mocks.NewMockOrderRepository(t),
		txOrderItem:        mocks.NewMockOrderItemRepository(t),
		txOrderAddr:        mocks.NewMockOrderAddressRepository(t),
		txHistory:          mocks.NewMockOrderStatusHistoryRepository(t),
		txPayment:          mocks.NewMockPaymentRepository(t),
		txShipment:         mocks.NewMockShipmentRepository(t),
		txAddress:          mocks.NewMockAddressRepository(t),
		txCart:             mocks.NewMockCartRepository(t),
		txVariant:          mocks.NewMockProductVariantRepository(t),
		txWallet:           mocks.NewMockWalletRepository(t),
		txCoupon:           mocks.NewMockCouponRepository(t),
		txCouponUse:        mocks.NewMockCouponUsageRepository(t),
		txUser:             mocks.NewMockUserRepository(t),
		txNotification:     mocks.NewMockNotificationRepository(t),
		txReturnRequest:    mocks.NewMockReturnRequestRepository(t),
		txReferral:         mocks.NewMockReferralRepository(t),
		txReferralSettings: mocks.NewMockReferralSettingsRepository(t),
		txWalletTopUp:      mocks.NewMockWalletTopUpRepository(t),
	}
}

func (f *orderFixture) expectTX(ctx context.Context) {
	f.uow.EXPECT().WithinTransaction(ctx, mock.AnythingOfType("func(interfaces.TransactionRepositories) error")).
		RunAndReturn(func(_ context.Context, fn func(interfaces.TransactionRepositories) error) error {
			return fn(interfaces.TransactionRepositories{
				User:               f.txUser,
				Notification:       f.txNotification,
				Order:              f.txOrder,
				OrderItem:          f.txOrderItem,
				OrderAddress:       f.txOrderAddr,
				OrderStatusHistory: f.txHistory,
				Shipment:           f.txShipment,
				ReturnRequest:      f.txReturnRequest,
				Referral:           f.txReferral,
				ReferralSettings:   f.txReferralSettings,
				Wallet:             f.txWallet,
				Payment:            f.txPayment,
				Coupon:             f.txCoupon,
				CouponUsage:        f.txCouponUse,
				WalletTopUp:        f.txWalletTopUp,
				Address:            f.txAddress,
				ProductVariant:     f.txVariant,
				Cart:               f.txCart,
			})
		}).Once()
}

func TestOrderUseCase_Queries(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()

	t.Run("GetByID and GetByOrderNumber", func(t *testing.T) {
		f := newOrderFixture(t)
		_, err := f.uc.GetByID(ctx, uuid.Nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidOrderID)

		ord := &models.Order{ID: orderID, OrderNumber: "ORD-1"}
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(ord, nil).Once()
		got, err := f.uc.GetByID(ctx, orderID)
		require.NoError(t, err)
		assert.Equal(t, ord, got)

		_, err = f.uc.GetByOrderNumber(ctx, "   ")
		assert.ErrorIs(t, err, domainerrors.ErrInvalidOrderNumber)

		f.orderRepo.EXPECT().GetByOrderNumber(ctx, "ORD-1").Return(ord, nil).Once()
		got, err = f.uc.GetByOrderNumber(ctx, " ORD-1 ")
		require.NoError(t, err)
		assert.Equal(t, ord, got)
	})
	t.Run("ListByUserID and List", func(t *testing.T) {
		f := newOrderFixture(t)
		_, _, err := f.uc.ListByUserID(ctx, 0, 10, 1, nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidUserID)

		list := []models.Order{{ID: orderID}}
		f.orderRepo.EXPECT().ListByUserID(mock.Anything, uint(7), 10, 1, (*string)(nil)).Return(list, int64(1), nil).Once()
		got, total, err := f.uc.ListByUserID(ctx, 7, 0, 0, nil)
		require.NoError(t, err)
		assert.Equal(t, list, got)
		assert.Equal(t, int64(1), total)

		f.orderRepo.EXPECT().List(mock.Anything, (*uint)(nil), "ORD", 10, 1, (*string)(nil), (*string)(nil), (*string)(nil)).
			Return(list, int64(1), nil).Once()
		got, total, err = f.uc.List(ctx, nil, "ORD", 10, 1, nil, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, list, got)
		assert.Equal(t, int64(1), total)
	})
}

func TestOrderUseCase_PlaceOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("validation errors", func(t *testing.T) {
		f := newOrderFixture(t)
		_, err := f.uc.PlaceOrder(ctx, 0, []uint{1}, 1, "cod", nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidUserID)

		_, err = f.uc.PlaceOrder(ctx, 1, nil, 1, "cod", nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, err = f.uc.PlaceOrder(ctx, 1, []uint{1}, 0, "cod", nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidAddress)

		_, err = f.uc.PlaceOrder(ctx, 1, []uint{1}, 1, "bitcoin", nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)
	})

	setupCartTX := func(f *orderFixture) {
		addr := &models.Address{
			ID: 5, UserID: 1, Name: "Niyas", Phone: "9876543210",
			AddressLine1: "Street 1", City: "Kochi", State: "KL", PostalCode: "682001", Country: "India",
		}
		variant := &models.ProductVariant{
			ID: 10, ProductID: 2, SKU: "SKU1", SellingPrice: 500, Stock: 10, IsActive: true,
			Product: models.Product{ID: 2, Name: "Shirt", IsActive: true, IsListed: true, DeletedAt: nil,
				Brand:    models.Brand{IsActive: true, DeletedAt: nil},
				Category: models.Category{IsActive: true, DeletedAt: nil},
			},
		}
		cart := &models.Cart{ID: 3, UserID: 1}
		items := []models.CartItem{{ID: 100, CartID: 3, VariantID: 10, Quantity: 2, Variant: *variant}}

		f.txAddress.EXPECT().FindByID(mock.Anything, uint(1), uint(5)).Return(addr, nil).Once()
		f.txCart.EXPECT().GetCartByUserID(mock.Anything, int64(1)).Return(cart, nil).Once()
		f.txCart.EXPECT().GetCartItemsByIDs(mock.Anything, int64(3), []uint{100}).Return(items, nil).Once()
		f.txVariant.EXPECT().GetByID(mock.Anything, int64(10)).Return(variant, nil).Maybe()
		f.txVariant.EXPECT().DecreaseStock(mock.Anything, int64(10), 2).Return(nil).Once()
		f.txOrder.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.Order")).Return(nil).Once()
		f.txOrderItem.EXPECT().CreateMany(mock.Anything, mock.AnythingOfType("[]models.OrderItem")).Return(nil).Maybe()
		f.txOrderItem.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.OrderItem")).Return(nil).Maybe()
		f.txOrderAddr.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.OrderAddress")).Return(nil).Once()
		f.txHistory.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.OrderStatusHistory")).Return(nil).Maybe()
		f.txPayment.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.Payment")).Return(nil).Once()
		f.txShipment.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.Shipment")).Return(nil).Maybe()
		f.txCart.EXPECT().RemoveCartItems(mock.Anything, int64(3), []uint{100}).Return(nil).Once()
		f.txUser.EXPECT().
			ListAdminIDs(mock.Anything).
			Return([]uint{1, 2}, nil).
			Once()
		f.txNotification.EXPECT().
			Create(mock.Anything, mock.AnythingOfType("*models.Notification")).
			Return(nil).
			Once()
		f.txNotification.EXPECT().
			CreateRecipients(mock.Anything, mock.AnythingOfType("[]models.NotificationRecipient")).
			Return(nil).
			Once()
	}

	t.Run("COD success", func(t *testing.T) {
		f := newOrderFixture(t)
		f.expectTX(ctx)
		setupCartTX(f)

		got, err := f.uc.PlaceOrder(ctx, 1, []uint{100}, 5, "cod", nil)
		require.NoError(t, err)
		// t.Errorf("%v", err)
		assert.Equal(t, "confirmed", got.OrderStatus)
		assert.Equal(t, "pending", got.PaymentStatus)
	})

	t.Run("online success", func(t *testing.T) {
		f := newOrderFixture(t)
		f.expectTX(ctx)
		setupCartTX(f)

		got, err := f.uc.PlaceOrder(ctx, 1, []uint{100}, 5, "online", nil)
		require.NoError(t, err)
		assert.Equal(t, "pending", got.OrderStatus)
		assert.Equal(t, "pending", got.PaymentStatus)
	})

	t.Run("wallet payment sufficient and insufficient balance", func(t *testing.T) {
		f := newOrderFixture(t)
		f.expectTX(ctx)
		f.txAddress.EXPECT().FindByID(ctx, uint(1), uint(5)).Return(&models.Address{
			ID: 5, UserID: 1, Name: "Niyas", Phone: "9876543210",
			AddressLine1: "Street 1", City: "Kochi", State: "KL", PostalCode: "682001", Country: "India",
		}, nil).Once()
		f.txCart.EXPECT().GetCartByUserID(ctx, int64(1)).Return(&models.Cart{ID: 3, UserID: 1}, nil).Once()
		variant := &models.ProductVariant{
			ID: 10, ProductID: 2, SKU: "SKU1", SellingPrice: 500, Stock: 10, IsActive: true,
			Product: models.Product{
				ID: 2, Name: "Shirt", IsActive: true, IsListed: true,
				Brand:    models.Brand{IsActive: true},
				Category: models.Category{IsActive: true},
			},
		}
		f.txCart.EXPECT().GetCartItemsByIDs(ctx, int64(3), []uint{100}).Return([]models.CartItem{
			{ID: 100, CartID: 3, VariantID: 10, Quantity: 2, Variant: *variant},
		}, nil).Once()
		f.txVariant.EXPECT().GetByID(ctx, int64(10)).Return(variant, nil).Maybe()
		f.txVariant.EXPECT().DecreaseStock(ctx, int64(10), 2).Return(nil).Maybe()
		f.txWallet.EXPECT().GetByUserIDForUpdate(ctx, uint(1)).Return(&models.Wallet{ID: 1, Balance: 50}, nil).Once()

		_, err := f.uc.PlaceOrder(ctx, 1, []uint{100}, 5, "wallet", nil)
		assert.ErrorIs(t, err, domainerrors.ErrInsufficientBalance)

		f2 := newOrderFixture(t)
		f2.expectTX(ctx)
		setupCartTX(f2)
		f2.txWallet.EXPECT().GetByUserIDForUpdate(ctx, uint(1)).Return(&models.Wallet{ID: 1, Balance: 5000}, nil).Once()
		f2.txWallet.EXPECT().Update(ctx, mock.AnythingOfType("*models.Wallet")).Return(nil).Once()
		f2.txWallet.EXPECT().CreateTransaction(ctx, mock.AnythingOfType("*models.WalletTransaction")).Return(nil).Once()

		got, err := f2.uc.PlaceOrder(ctx, 1, []uint{100}, 5, "wallet", nil)
		require.NoError(t, err)
		assert.Equal(t, "confirmed", got.OrderStatus)
		assert.Equal(t, "paid", got.PaymentStatus)
	})

	t.Run("coupon discount branch without expecting shadowed CouponUsage.Create", func(t *testing.T) {
		f := newOrderFixture(t)
		f.expectTX(ctx)
		setupCartTX(f)
		code := "SAVE10"
		limit, userLimit := 10, 1
		coupon := &models.Coupon{
			ID: uuid.New(), Code: "SAVE10", DiscountType: "fixed", DiscountValue: 50,
			MinimumOrderAmount: 100, StartAt: time.Now().Add(-time.Hour), ExpiresAt: time.Now().Add(time.Hour), IsActive: true,
			UsageLimit: &limit, UsageLimitPerUser: &userLimit,
		}
		f.txCoupon.EXPECT().GetByCodeForUpdate(ctx, "SAVE10").Return(coupon, nil).Once()
		f.txCouponUse.EXPECT().CountByCouponID(ctx, coupon.ID).Return(int64(0), nil).Once()
		f.txCouponUse.EXPECT().CountByCouponAndUser(ctx, coupon.ID, uint(1)).Return(int64(0), nil).Once()

		got, err := f.uc.PlaceOrder(ctx, 1, []uint{100}, 5, "cod", &code)
		require.NoError(t, err)
		assert.Equal(t, 50.0, got.DiscountAmount)
	})
}

func TestOrderUseCase_StatusCancelAddressInvoicePaymentExpireRefund(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()

	t.Run("UpdateOrderStatus", func(t *testing.T) {
		f := newOrderFixture(t)
		assert.ErrorIs(t, f.uc.UpdateOrderStatus(ctx, uuid.Nil, "confirmed", nil), domainerrors.ErrInvalidOrderID)
		assert.ErrorIs(t, f.uc.UpdateOrderStatus(ctx, orderID, "invalid", nil), domainerrors.ErrInvalidOrderStatus)

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{ID: orderID, OrderStatus: "pending"}, nil).Once()
		f.orderRepo.EXPECT().UpdateOrderStatus(ctx, orderID, "confirmed", (*string)(nil)).Return(nil).Once()

		require.NoError(t, f.uc.UpdateOrderStatus(ctx, orderID, "confirmed", nil))
	})

	t.Run("CancelOrder", func(t *testing.T) {
		f := newOrderFixture(t)
		assert.ErrorIs(t, f.uc.CancelOrder(ctx, 0, orderID), domainerrors.ErrInvalidUserID)
		assert.ErrorIs(t, f.uc.CancelOrder(ctx, 1, uuid.Nil), domainerrors.ErrInvalidOrderID)

		f.expectTX(ctx)
		ord := &models.Order{
			ID: orderID, UserID: 1, OrderStatus: "confirmed", PaymentStatus: "paid", TotalAmount: 500,
			OrderNumber: "ORD-CANCEL-1",
			Items:       []models.OrderItem{{ID: 1, VariantID: 10, Quantity: 2}},
		}
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(ord, nil).Once()
		f.txOrderItem.EXPECT().ListByOrderID(ctx, orderID).Return(ord.Items, nil).Maybe()
		f.txVariant.EXPECT().IncreaseStock(ctx, int64(10), 2).Return(nil).Once()
		f.txCouponUse.EXPECT().CancelByOrderID(ctx, orderID).Return(nil).Maybe()
		f.txOrder.EXPECT().Update(ctx, mock.AnythingOfType("*models.Order")).Return(nil).Maybe()
		f.txHistory.EXPECT().Create(ctx, mock.AnythingOfType("*models.OrderStatusHistory")).Return(nil).Maybe()
		f.txNotification.EXPECT().Create(ctx, mock.AnythingOfType("*models.Notification")).Return(nil).Once()
		f.txUser.EXPECT().ListAdminIDs(ctx).Return([]uint{1}, nil).Once()
		f.txNotification.EXPECT().CreateRecipients(ctx, mock.AnythingOfType("[]models.NotificationRecipient")).Return(nil).Once()

		require.NoError(t, f.uc.CancelOrder(ctx, 1, orderID))
	})

	t.Run("UpdateOrderAddress", func(t *testing.T) {
		f := newOrderFixture(t)
		addr := models.OrderAddress{
			Name: "Niyas", Phone: "9876543210", AddressLine1: "Line 1",
			City: "Kochi", State: "KL", PostalCode: "682001", Country: "India",
		}
		assert.ErrorIs(t, f.uc.UpdateOrderAddress(ctx, 0, orderID, addr), domainerrors.ErrInvalidUserID)
		assert.ErrorIs(t, f.uc.UpdateOrderAddress(ctx, 1, uuid.Nil, addr), domainerrors.ErrInvalidOrderID)

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 1, OrderStatus: "pending"}, nil).Maybe()
		f.expectTX(ctx)
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 1, OrderStatus: "pending"}, nil).Maybe()
		f.txOrderAddr.EXPECT().GetByOrderID(ctx, orderID).Return(&models.OrderAddress{ID: 9, OrderID: orderID}, nil).Once()
		f.txOrderAddr.EXPECT().Update(ctx, mock.MatchedBy(func(a *models.OrderAddress) bool {
			return a.City == "Kochi"
		})).Return(nil).Once()

		require.NoError(t, f.uc.UpdateOrderAddress(ctx, 1, orderID, addr))
	})

	t.Run("GenerateInvoice", func(t *testing.T) {
		f := newOrderFixture(t)
		_, err := f.uc.GenerateInvoice(ctx, 0, orderID)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)
		_, err = f.uc.GenerateInvoice(ctx, 1, uuid.Nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		ord := &models.Order{
			ID: orderID, UserID: 1,
			Items:   []models.OrderItem{{ID: 1, ProductName: "Shirt", Quantity: 1, Subtotal: 500}},
			Address: &models.OrderAddress{Name: "Niyas", City: "Kochi"},
		}
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(ord, nil).Once()
		f.invoiceGen.EXPECT().Generate(ord).Return([]byte("%PDF-1.4"), nil).Once()

		pdf, err := f.uc.GenerateInvoice(ctx, 1, orderID)
		require.NoError(t, err)
		assert.Equal(t, []byte("%PDF-1.4"), pdf)
	})

	t.Run("UpdateCODPaymentStatus ExpirePendingOrder ProcessWalletRefund", func(t *testing.T) {
		f := newOrderFixture(t)
		assert.ErrorIs(t, f.uc.UpdateCODPaymentStatus(ctx, uuid.Nil, "paid"), domainerrors.ErrInvalidOrderID)

		f.expectTX(ctx)
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(&models.Order{
			ID: orderID, UserID: 1, OrderStatus: "confirmed", PaymentStatus: "pending",
		}, nil).Once()
		f.txPayment.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{
			{OrderID: orderID, PaymentMethod: "cod", PaymentStatus: "pending"},
		}, nil).Once()
		f.txPayment.EXPECT().Update(ctx, mock.AnythingOfType("*models.Payment")).Return(nil).Once()
		f.txOrder.EXPECT().Update(ctx, mock.AnythingOfType("*models.Order")).Return(nil).Once()
		require.NoError(t, f.uc.UpdateCODPaymentStatus(ctx, orderID, "paid"))

		f2 := newOrderFixture(t)
		assert.ErrorIs(t, f2.uc.ExpirePendingOrder(ctx, uuid.Nil), domainerrors.ErrInvalidOrderID)
		f2.expectTX(ctx)
		f2.txOrder.EXPECT().LockByID(ctx, orderID).Return(&models.Order{
			ID: orderID, UserID: 1, OrderStatus: "pending", PaymentStatus: "pending",
			PlacedAt: time.Now().Add(-15 * time.Minute),
			Items:    []models.OrderItem{{ID: 1, VariantID: 10, Quantity: 2}},
		}, nil).Once()
		f2.txPayment.EXPECT().GetByOrderID(ctx, orderID).Return([]models.Payment{
			{OrderID: orderID, PaymentMethod: "online", PaymentStatus: "pending"},
		}, nil).Once()
		f2.txOrderItem.EXPECT().ListByOrderID(ctx, orderID).Return([]models.OrderItem{
			{ID: 1, VariantID: 10, Quantity: 2},
		}, nil).Once()
		f2.txVariant.EXPECT().IncreaseStock(ctx, int64(10), 2).Return(nil).Once()
		f2.txCouponUse.EXPECT().CancelByOrderID(ctx, orderID).Return(nil).Maybe()
		f2.txPayment.EXPECT().Update(ctx, mock.AnythingOfType("*models.Payment")).Return(nil).Once()
		f2.txOrder.EXPECT().Update(ctx, mock.AnythingOfType("*models.Order")).Return(nil).Once()
		f2.txHistory.EXPECT().Create(ctx, mock.AnythingOfType("*models.OrderStatusHistory")).Return(nil).Maybe()
		require.NoError(t, f2.uc.ExpirePendingOrder(ctx, orderID))

		f3 := newOrderFixture(t)
		assert.ErrorIs(t, f3.uc.ProcessWalletRefund(ctx, 1, uuid.Nil), domainerrors.ErrInvalidOrderID)
		f3.expectTX(ctx)
		f3.txOrder.EXPECT().LockByID(ctx, orderID).Return(&models.Order{
			ID: orderID, UserID: 7, OrderStatus: "cancelled", PaymentStatus: "paid", TotalAmount: 300,
			OrderNumber: "ORD-REFUND-1",
		}, nil).Once()
		f3.txPayment.EXPECT().GetLatestByOrderIDForUpdate(ctx, orderID).Return(&models.Payment{
			OrderID: orderID, PaymentStatus: "paid", Amount: 300,
		}, nil).Once()
		f3.txWallet.EXPECT().GetByUserIDForUpdate(ctx, uint(7)).Return(&models.Wallet{ID: 7, Balance: 50}, nil).Once()
		f3.txWallet.EXPECT().Update(ctx, mock.AnythingOfType("*models.Wallet")).Return(nil).Once()
		f3.txWallet.EXPECT().CreateTransaction(ctx, mock.AnythingOfType("*models.WalletTransaction")).Return(nil).Once()
		f3.txOrder.EXPECT().Update(ctx, mock.AnythingOfType("*models.Order")).Return(nil).Once()
		require.NoError(t, f3.uc.ProcessWalletRefund(ctx, 1, orderID))
	})
}
