package return_request

import (
	"context"
	"io"
	"log/slog"
	"testing"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIsValidReturnStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "requested status",
			status: ReturnStatusRequested,
			want:   true,
		},
		{
			name:   "approved status",
			status: ReturnStatusApproved,
			want:   true,
		},
		{
			name:   "rejected status",
			status: ReturnStatusRejected,
			want:   true,
		},
		{
			name:   "pickup scheduled status",
			status: ReturnStatusPickupScheduled,
			want:   true,
		},
		{
			name:   "in transit status",
			status: ReturnStatusInTransit,
			want:   true,
		},
		{
			name:   "received status",
			status: ReturnStatusReceived,
			want:   true,
		},
		{
			name:   "inspection pending status",
			status: ReturnStatusInspectionPending,
			want:   true,
		},
		{
			name:   "inspection passed status",
			status: ReturnStatusInspectionPassed,
			want:   true,
		},
		{
			name:   "inspection failed status",
			status: ReturnStatusInspectionFailed,
			want:   true,
		},
		{
			name:   "closed status",
			status: ReturnStatusClosed,
			want:   true,
		},

		// Invalid statuses
		{
			name:   "unknown status",
			status: "unknown",
			want:   false,
		},
		{
			name:   "empty status",
			status: "",
			want:   false,
		},
		{
			name:   "incorrect status casing",
			status: "REQUESTED",
			want:   false,
		},
		{
			name:   "status with whitespace",
			status: " requested ",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidReturnStatus(tt.status)

			if got != tt.want {
				t.Errorf(
					"isValidReturnStatus(%q) = %v; want %v",
					tt.status,
					got,
					tt.want,
				)
			}
		})
	}
}
func TestIsValidReturnTransition(t *testing.T) {
	tests := []struct {
		name    string
		current string
		next    string
		want    bool
	}{
		// Requested transitions
		{
			name:    "requested to approved",
			current: ReturnStatusRequested,
			next:    ReturnStatusApproved,
			want:    true,
		},
		{
			name:    "requested to rejected",
			current: ReturnStatusRequested,
			next:    ReturnStatusRejected,
			want:    true,
		},

		// Approved transitions
		{
			name:    "approved to pickup scheduled",
			current: ReturnStatusApproved,
			next:    ReturnStatusPickupScheduled,
			want:    true,
		},
		{
			name:    "approved to closed",
			current: ReturnStatusApproved,
			next:    ReturnStatusClosed,
			want:    true,
		},

		// Pickup scheduled transitions
		{
			name:    "pickup scheduled to in transit",
			current: ReturnStatusPickupScheduled,
			next:    ReturnStatusInTransit,
			want:    true,
		},
		{
			name:    "pickup scheduled to closed",
			current: ReturnStatusPickupScheduled,
			next:    ReturnStatusClosed,
			want:    true,
		},

		// In transit transitions
		{
			name:    "in transit to received",
			current: ReturnStatusInTransit,
			next:    ReturnStatusReceived,
			want:    true,
		},

		// Received transitions
		{
			name:    "received to inspection pending",
			current: ReturnStatusReceived,
			next:    ReturnStatusInspectionPending,
			want:    true,
		},

		// Inspection pending transitions
		{
			name:    "inspection pending to inspection passed",
			current: ReturnStatusInspectionPending,
			next:    ReturnStatusInspectionPassed,
			want:    true,
		},
		{
			name:    "inspection pending to inspection failed",
			current: ReturnStatusInspectionPending,
			next:    ReturnStatusInspectionFailed,
			want:    true,
		},

		// Inspection result transitions
		{
			name:    "inspection passed to closed",
			current: ReturnStatusInspectionPassed,
			next:    ReturnStatusClosed,
			want:    true,
		},
		{
			name:    "inspection failed to closed",
			current: ReturnStatusInspectionFailed,
			next:    ReturnStatusClosed,
			want:    true,
		},

		// Invalid transitions
		{
			name:    "requested to received",
			current: ReturnStatusRequested,
			next:    ReturnStatusReceived,
			want:    false,
		},
		{
			name:    "requested to closed",
			current: ReturnStatusRequested,
			next:    ReturnStatusClosed,
			want:    false,
		},
		{
			name:    "approved to rejected",
			current: ReturnStatusApproved,
			next:    ReturnStatusRejected,
			want:    false,
		},
		{
			name:    "in transit to closed",
			current: ReturnStatusInTransit,
			next:    ReturnStatusClosed,
			want:    false,
		},
		{
			name:    "received to inspection passed",
			current: ReturnStatusReceived,
			next:    ReturnStatusInspectionPassed,
			want:    false,
		},
		{
			name:    "inspection passed to inspection failed",
			current: ReturnStatusInspectionPassed,
			next:    ReturnStatusInspectionFailed,
			want:    false,
		},

		// Edge cases
		{
			name:    "same status transition",
			current: ReturnStatusRequested,
			next:    ReturnStatusRequested,
			want:    false,
		},
		{
			name:    "unknown current status",
			current: "unknown",
			next:    ReturnStatusApproved,
			want:    false,
		},
		{
			name:    "unknown next status",
			current: ReturnStatusRequested,
			next:    "unknown",
			want:    false,
		},
		{
			name:    "both statuses empty",
			current: "",
			next:    "",
			want:    false,
		},
		{
			name:    "empty current status",
			current: "",
			next:    ReturnStatusApproved,
			want:    false,
		},
		{
			name:    "empty next status",
			current: ReturnStatusRequested,
			next:    "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidReturnTransition(tt.current, tt.next)

			if got != tt.want {
				t.Errorf(
					"isValidReturnTransition(%q, %q) = %v; want %v",
					tt.current,
					tt.next,
					got,
					tt.want,
				)
			}
		})
	}
}

type returnFixture struct {
	uc                 *ReturnRequestUseCase
	retRepo            *mocks.MockReturnRequestRepository
	orderRepo          *mocks.MockOrderRepository
	itemRepo           *mocks.MockOrderItemRepository
	variantRepo        *mocks.MockProductVariantRepository
	shipRepo           *mocks.MockShipmentRepository
	uow                *mocks.MockUnitOfWork
	txRet              *mocks.MockReturnRequestRepository
	txOrder            *mocks.MockOrderRepository
	txItem             *mocks.MockOrderItemRepository
	txAddr             *mocks.MockOrderAddressRepository
	txHistory          *mocks.MockOrderStatusHistoryRepository
	txVariant          *mocks.MockProductVariantRepository
	txShip             *mocks.MockShipmentRepository
	txWallet           *mocks.MockWalletRepository
	txUser             *mocks.MockUserRepository
	txNotification     *mocks.MockNotificationRepository
	txReferral         *mocks.MockReferralRepository
	txReferralSettings *mocks.MockReferralSettingsRepository
	txPayment          *mocks.MockPaymentRepository
	txCoupon           *mocks.MockCouponRepository
	txCouponUsage      *mocks.MockCouponUsageRepository
	txWalletTopUp      *mocks.MockWalletTopUpRepository
	txAddress          *mocks.MockAddressRepository
	txCart             *mocks.MockCartRepository
}

func newReturnFixture(t *testing.T) *returnFixture {
	t.Helper()
	retRepo := mocks.NewMockReturnRequestRepository(t)
	orderRepo := mocks.NewMockOrderRepository(t)
	itemRepo := mocks.NewMockOrderItemRepository(t)
	variantRepo := mocks.NewMockProductVariantRepository(t)
	shipRepo := mocks.NewMockShipmentRepository(t)
	uow := mocks.NewMockUnitOfWork(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &returnFixture{
		uc:                 NewReturnRequestUseCase(retRepo, orderRepo, itemRepo, variantRepo, shipRepo, uow, log),
		retRepo:            retRepo,
		orderRepo:          orderRepo,
		itemRepo:           itemRepo,
		variantRepo:        variantRepo,
		shipRepo:           shipRepo,
		uow:                uow,
		txRet:              mocks.NewMockReturnRequestRepository(t),
		txOrder:            mocks.NewMockOrderRepository(t),
		txItem:             mocks.NewMockOrderItemRepository(t),
		txAddr:             mocks.NewMockOrderAddressRepository(t),
		txHistory:          mocks.NewMockOrderStatusHistoryRepository(t),
		txVariant:          mocks.NewMockProductVariantRepository(t),
		txShip:             mocks.NewMockShipmentRepository(t),
		txWallet:           mocks.NewMockWalletRepository(t),
		txUser:             mocks.NewMockUserRepository(t),
		txNotification:     mocks.NewMockNotificationRepository(t),
		txReferral:         mocks.NewMockReferralRepository(t),
		txReferralSettings: mocks.NewMockReferralSettingsRepository(t),
		txPayment:          mocks.NewMockPaymentRepository(t),
		txCoupon:           mocks.NewMockCouponRepository(t),
		txCouponUsage:      mocks.NewMockCouponUsageRepository(t),
		txWalletTopUp:      mocks.NewMockWalletTopUpRepository(t),
		txAddress:          mocks.NewMockAddressRepository(t),
		txCart:             mocks.NewMockCartRepository(t),
	}
}

func (f *returnFixture) expectTX(ctx context.Context) {
	f.uow.EXPECT().WithinTransaction(ctx, mock.AnythingOfType("func(interfaces.TransactionRepositories) error")).
		RunAndReturn(func(_ context.Context, fn func(interfaces.TransactionRepositories) error) error {
			return fn(interfaces.TransactionRepositories{
				User:               f.txUser,
				Notification:       f.txNotification,
				Order:              f.txOrder,
				OrderItem:          f.txItem,
				OrderAddress:       f.txAddr,
				OrderStatusHistory: f.txHistory,
				Shipment:           f.txShip,
				ReturnRequest:      f.txRet,
				Referral:           f.txReferral,
				ReferralSettings:   f.txReferralSettings,
				Wallet:             f.txWallet,
				Payment:            f.txPayment,
				Coupon:             f.txCoupon,
				CouponUsage:        f.txCouponUsage,
				WalletTopUp:        f.txWalletTopUp,
				Address:            f.txAddress,
				ProductVariant:     f.txVariant,
				Cart:               f.txCart,
			})
		}).Once()
}

func TestReturnRequestUseCase_CreateAndQueries(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()

	t.Run("Create validation and return flow", func(t *testing.T) {
		f := newReturnFixture(t)
		_, err := f.uc.Create(ctx, 0, orderID, 1, RequestTypeReturn, nil, 1, "Damaged")
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, err = f.uc.Create(ctx, 1, uuid.Nil, 1, RequestTypeReturn, nil, 1, "Damaged")
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, err = f.uc.Create(ctx, 1, orderID, 1, "invalid_type", nil, 1, "Damaged")
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, err = f.uc.Create(ctx, 1, orderID, 1, RequestTypeReturn, nil, 0, "Damaged")
		assert.ErrorIs(t, err, domainerrors.ErrInvalidQuantity)

		_, err = f.uc.Create(ctx, 1, orderID, 1, RequestTypeReturn, nil, 1, "   ")
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{
			ID: orderID, UserID: 1, OrderStatus: "delivered", PaymentStatus: "paid", OrderNumber: "ORD-RET-1",
		}, nil).Once()
		f.itemRepo.EXPECT().GetByID(ctx, uint(10)).Return(&models.OrderItem{
			ID: 10, OrderID: orderID, VariantID: 5, Quantity: 2, UnitPrice: 300,
		}, nil).Maybe()
		f.expectTX(ctx)
		f.txRet.EXPECT().Create(ctx, mock.MatchedBy(func(r *models.ReturnRequest) bool {
			return r.OrderID == orderID && r.OrderItemID == 10 && r.Quantity == 1 && r.RequestType == RequestTypeReturn
		})).Return(nil).Once()
		f.txUser.EXPECT().ListAdminIDs(ctx).Return([]uint{1}, nil).Once()
		f.txNotification.EXPECT().Create(ctx, mock.AnythingOfType("*models.Notification")).Return(nil).Once()
		f.txNotification.EXPECT().CreateRecipients(ctx, mock.AnythingOfType("[]models.NotificationRecipient")).Return(nil).Once()

		got, err := f.uc.Create(ctx, 1, orderID, 10, RequestTypeReturn, nil, 1, "Wrong size")
		require.NoError(t, err)
		assert.Equal(t, ReturnStatusRequested, got.Status)
	})

	t.Run("GetByID ListByUserID List GetCustomerShipments", func(t *testing.T) {
		f := newReturnFixture(t)
		_, err := f.uc.GetByID(ctx, 0, 1)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		req := &models.ReturnRequest{ID: 9, OrderID: orderID}
		f.retRepo.EXPECT().GetByIDAndUserID(ctx, uint(9), uint(1)).Return(req, nil).Once()
		got, err := f.uc.GetByID(ctx, 9, 1)
		require.NoError(t, err)
		assert.Equal(t, req, got)

		_, _, err = f.uc.ListByUserID(ctx, 0, 10, 1)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		f.retRepo.EXPECT().ListByUserID(ctx, uint(1), 10, 1).Return([]models.ReturnRequest{*req}, int64(1), nil).Once()
		list, total, err := f.uc.ListByUserID(ctx, 1, 0, 0)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, int64(1), total)

		badStatus := "bogus"
		_, _, err = f.uc.List(ctx, nil, &badStatus, 10, 1)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		f.retRepo.EXPECT().List(ctx, &orderID, (*string)(nil), 10, 1).Return([]models.ReturnRequest{*req}, int64(1), nil).Once()
		list, total, err = f.uc.List(ctx, &orderID, nil, 0, 0)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, int64(1), total)

		ships := []models.Shipment{{ID: 2}}
		f.retRepo.EXPECT().GetByIDAndUserID(ctx, uint(9), uint(1)).Return(req, nil).Once()
		f.shipRepo.EXPECT().GetByReturnRequestID(ctx, uint(9)).Return(ships, nil).Once()
		gotShips, err := f.uc.GetCustomerShipments(ctx, 9, 1)
		require.NoError(t, err)
		assert.Equal(t, ships, gotShips)
	})
}

func TestReturnRequestUseCase_UpdateStatusAndFulfillment(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()

	t.Run("UpdateStatus pickup_scheduled and inspection_passed return refund", func(t *testing.T) {
		f := newReturnFixture(t)
		_, err := f.uc.UpdateStatus(ctx, 0, ReturnStatusApproved, nil, nil, nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		courier, tracking := "Delhivery", "TRK-99"
		f.expectTX(ctx)
		f.txRet.EXPECT().LockByID(ctx, uint(11)).Return(&models.ReturnRequest{
			ID: 11, OrderID: orderID, OrderItemID: 1, RequestType: RequestTypeReturn, Status: ReturnStatusApproved,
		}, nil).Once()
		retShipID := uint(11)
		f.txShip.EXPECT().GetByReturnRequestID(ctx, uint(11)).Return([]models.Shipment{
			{ID: 1, OrderID: orderID, ReturnRequestID: &retShipID, Type: models.ShipmentTypeReturn},
		}, nil).Once()
		f.txShip.EXPECT().Update(ctx, mock.AnythingOfType("*models.Shipment")).Return(nil).Once()
		f.txRet.EXPECT().Update(ctx, mock.AnythingOfType("*models.ReturnRequest")).Return(nil).Once()

		got, err := f.uc.UpdateStatus(ctx, 11, ReturnStatusPickupScheduled, nil, &courier, &tracking)
		require.NoError(t, err)
		assert.Equal(t, ReturnStatusPickupScheduled, got.Status)

		f2 := newReturnFixture(t)
		f2.expectTX(ctx)
		f2.txRet.EXPECT().LockByID(ctx, uint(11)).Return(&models.ReturnRequest{
			ID: 11, OrderID: orderID, OrderItemID: 1, RequestType: RequestTypeReturn, Quantity: 1,
			Status: ReturnStatusInspectionPending,
			Order:  models.Order{ID: orderID, UserID: 5, PaymentStatus: "paid", TotalAmount: 400},
			OrderItem: models.OrderItem{
				ID: 1, VariantID: 20, UnitPrice: 400, Quantity: 1, Subtotal: 400,
			},
		}, nil).Once()
		f2.txOrder.EXPECT().GetByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 5, PaymentStatus: "paid", TotalAmount: 400}, nil).Maybe()
		f2.txOrder.EXPECT().LockByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 5, PaymentStatus: "paid", TotalAmount: 400}, nil).Maybe()
		f2.txItem.EXPECT().GetByID(ctx, uint(1)).Return(&models.OrderItem{ID: 1, VariantID: 20, UnitPrice: 400, Quantity: 1, Subtotal: 400}, nil).Maybe()
		f2.txWallet.EXPECT().GetByUserIDForUpdate(ctx, uint(5)).Return(&models.Wallet{ID: 5, Balance: 100}, nil).Once()
		f2.txWallet.EXPECT().Update(ctx, mock.AnythingOfType("*models.Wallet")).Return(nil).Once()
		f2.txWallet.EXPECT().CreateTransaction(ctx, mock.AnythingOfType("*models.WalletTransaction")).Return(nil).Once()
		f2.txVariant.EXPECT().IncreaseStock(ctx, int64(20), 1).Return(nil).Maybe()
		f2.txOrder.EXPECT().Update(ctx, mock.AnythingOfType("*models.Order")).Return(nil).Maybe()
		f2.txRet.EXPECT().Update(ctx, mock.AnythingOfType("*models.ReturnRequest")).Return(nil).Once()

		got2, err := f2.uc.UpdateStatus(ctx, 11, ReturnStatusInspectionPassed, nil, nil, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, got2.Status)
	})

	t.Run("CreateExchangeOrder", func(t *testing.T) {
		f := newReturnFixture(t)
		_, err := f.uc.CreateExchangeOrder(ctx, 0)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		reqVarID := uint(30)
		f.expectTX(ctx)
		f.txRet.EXPECT().LockByID(ctx, uint(12)).Return(&models.ReturnRequest{
			ID:                 12,
			OrderID:            orderID,
			OrderItemID:        2,
			RequestType:        RequestTypeExchange,
			RequestedVariantID: &reqVarID,
			Quantity:           1,
			Status:             ReturnStatusInspectionPassed,
		}, nil).Once()
		f.txOrder.EXPECT().GetByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 5}, nil).Maybe()
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 5}, nil).Maybe()
		f.txItem.EXPECT().GetByID(ctx, uint(2)).Return(&models.OrderItem{
			ID: 2, OrderID: orderID, VariantID: 20, ProductName: "Shirt", UnitPrice: 500, Quantity: 1, Subtotal: 500,
		}, nil).Maybe()
		origVariant := &models.ProductVariant{
			ID: 20, ProductID: 1, SKU: "SKU-20", SellingPrice: 500, Stock: 5, IsActive: true,
			Product: models.Product{
				ID: 1, Name: "Shirt", IsActive: true, IsListed: true,
				Brand:    models.Brand{IsActive: true},
				Category: models.Category{IsActive: true},
			},
		}
		newVariant := &models.ProductVariant{
			ID: 30, ProductID: 1, SKU: "SKU-30", SellingPrice: 500, Stock: 5, IsActive: true,
			Product: models.Product{
				ID: 1, Name: "Shirt", IsActive: true, IsListed: true,
				Brand:    models.Brand{IsActive: true},
				Category: models.Category{IsActive: true},
			},
		}
		f.txVariant.EXPECT().GetByID(ctx, int64(20)).Return(origVariant, nil).Once()
		f.txVariant.EXPECT().GetByID(ctx, int64(30)).Return(newVariant, nil).Once()
		f.txVariant.EXPECT().DecreaseStock(ctx, int64(30), 1).Return(nil).Maybe()
		f.txAddr.EXPECT().GetByOrderID(ctx, orderID).Return(&models.OrderAddress{
			ID: 1, OrderID: orderID, Name: "Niyas", Phone: "9876543210",
			AddressLine1: "Line 1", City: "Kochi", State: "KL", PostalCode: "682001", Country: "India",
		}, nil).Maybe()
		f.txOrder.EXPECT().Create(ctx, mock.AnythingOfType("*models.Order")).Return(nil).Once()
		f.txItem.EXPECT().CreateMany(ctx, mock.AnythingOfType("[]models.OrderItem")).Return(nil).Maybe()
		f.txItem.EXPECT().Create(ctx, mock.AnythingOfType("*models.OrderItem")).Return(nil).Maybe()
		f.txAddr.EXPECT().Create(ctx, mock.AnythingOfType("*models.OrderAddress")).Return(nil).Maybe()
		f.txHistory.EXPECT().Create(ctx, mock.AnythingOfType("*models.OrderStatusHistory")).Return(nil).Maybe()
		f.txShip.EXPECT().Create(ctx, mock.MatchedBy(func(s *models.Shipment) bool {
			return s.Type == models.ShipmentTypeDelivery
		})).Return(nil).Once()
		f.txRet.EXPECT().Update(ctx, mock.AnythingOfType("*models.ReturnRequest")).Return(nil).Once()

		repOrder, err := f.uc.CreateExchangeOrder(ctx, 12)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, repOrder.ID)
	})

	t.Run("CreateReturnToCustomerShipment", func(t *testing.T) {
		f := newReturnFixture(t)
		_, err := f.uc.CreateReturnToCustomerShipment(ctx, 0, "BlueDart", "TRK-1")
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		_, err = f.uc.CreateReturnToCustomerShipment(ctx, 14, "", "TRK-1")
		assert.ErrorIs(t, err, domainerrors.ErrInvalidInput)

		f.expectTX(ctx)
		f.txRet.EXPECT().LockByID(ctx, uint(14)).Return(&models.ReturnRequest{
			ID: 14, OrderID: orderID, Status: ReturnStatusInspectionFailed,
		}, nil).Once()
		reqID14 := uint(14)
		f.txShip.EXPECT().GetByReturnRequestIDAndType(ctx, uint(14), models.ShipmentTypeReturn).
			Return(&models.Shipment{ID: 1, ReturnRequestID: &reqID14, Type: models.ShipmentTypeReturn, ShippingStatus: "delivered"}, nil).Once()
		f.txShip.EXPECT().GetByReturnRequestIDAndType(ctx, uint(14), models.ShipmentTypeReturnToCustomer).
			Return(nil, domainerrors.ErrRecordNotFound).Once()
		f.txShip.EXPECT().Create(ctx, mock.MatchedBy(func(s *models.Shipment) bool {
			return s.Type == models.ShipmentTypeReturnToCustomer && s.ReturnRequestID != nil && *s.ReturnRequestID == 14
		})).Return(nil).Once()

		got, err := f.uc.CreateReturnToCustomerShipment(ctx, 14, "BlueDart", "TRK-77")
		require.NoError(t, err)
		assert.Equal(t, models.ShipmentTypeReturnToCustomer, got.Type)
	})
}
