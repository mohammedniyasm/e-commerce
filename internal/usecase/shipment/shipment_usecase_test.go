package shipment

import (
	"context"
	"errors"
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
func TestIsValidReturnStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "requested",
			status: ReturnStatusRequested,
			want:   true,
		},
		{
			name:   "approved",
			status: ReturnStatusApproved,
			want:   true,
		},
		{
			name:   "rejected",
			status: ReturnStatusRejected,
			want:   true,
		},
		{
			name:   "pickup scheduled",
			status: ReturnStatusPickupScheduled,
			want:   true,
		},
		{
			name:   "in transit",
			status: ReturnStatusInTransit,
			want:   true,
		},
		{
			name:   "received",
			status: ReturnStatusReceived,
			want:   true,
		},
		{
			name:   "inspection pending",
			status: ReturnStatusInspectionPending,
			want:   true,
		},
		{
			name:   "inspection passed",
			status: ReturnStatusInspectionPassed,
			want:   true,
		},
		{
			name:   "inspection failed",
			status: ReturnStatusInspectionFailed,
			want:   true,
		},
		{
			name:   "closed",
			status: ReturnStatusClosed,
			want:   true,
		},
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
			name:   "incorrect casing",
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
				t.Errorf("isValidReturnStatus(%q) = %v; want %v",
					tt.status, got, tt.want)
			}
		})
	}
}
func TestIsValidShippingTransition(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		nextStatus    string
		want          bool
	}{
		// Pending transitions
		{
			name:          "pending to packed",
			currentStatus: "pending",
			nextStatus:    "packed",
			want:          true,
		},
		{
			name:          "pending to cancelled",
			currentStatus: "pending",
			nextStatus:    "cancelled",
			want:          true,
		},

		// Packed transitions
		{
			name:          "packed to ready to ship",
			currentStatus: "packed",
			nextStatus:    "ready_to_ship",
			want:          true,
		},
		{
			name:          "packed to cancelled",
			currentStatus: "packed",
			nextStatus:    "cancelled",
			want:          true,
		},

		// Ready to ship transitions
		{
			name:          "ready to ship to in transit",
			currentStatus: "ready_to_ship",
			nextStatus:    "in_transit",
			want:          true,
		},
		{
			name:          "ready to ship to cancelled",
			currentStatus: "ready_to_ship",
			nextStatus:    "cancelled",
			want:          true,
		},

		// In transit transitions
		{
			name:          "in transit to out for delivery",
			currentStatus: "in_transit",
			nextStatus:    "out_for_delivery",
			want:          true,
		},
		{
			name:          "in transit to failed",
			currentStatus: "in_transit",
			nextStatus:    "failed",
			want:          true,
		},
		{
			name:          "in transit to returned",
			currentStatus: "in_transit",
			nextStatus:    "returned",
			want:          true,
		},

		// Out for delivery transitions
		{
			name:          "out for delivery to delivered",
			currentStatus: "out_for_delivery",
			nextStatus:    "delivered",
			want:          true,
		},
		{
			name:          "out for delivery to failed",
			currentStatus: "out_for_delivery",
			nextStatus:    "failed",
			want:          true,
		},
		{
			name:          "out for delivery to returned",
			currentStatus: "out_for_delivery",
			nextStatus:    "returned",
			want:          true,
		},

		// Failed transitions
		{
			name:          "failed to in transit",
			currentStatus: "failed",
			nextStatus:    "in_transit",
			want:          true,
		},
		{
			name:          "failed to out for delivery",
			currentStatus: "failed",
			nextStatus:    "out_for_delivery",
			want:          true,
		},
		{
			name:          "failed to returned",
			currentStatus: "failed",
			nextStatus:    "returned",
			want:          true,
		},

		// Invalid transitions
		{
			name:          "pending to delivered",
			currentStatus: "pending",
			nextStatus:    "delivered",
			want:          false,
		},
		{
			name:          "packed to in transit",
			currentStatus: "packed",
			nextStatus:    "in_transit",
			want:          false,
		},
		{
			name:          "ready to ship to delivered",
			currentStatus: "ready_to_ship",
			nextStatus:    "delivered",
			want:          false,
		},
		{
			name:          "in transit to cancelled",
			currentStatus: "in_transit",
			nextStatus:    "cancelled",
			want:          false,
		},
		{
			name:          "out for delivery to cancelled",
			currentStatus: "out_for_delivery",
			nextStatus:    "cancelled",
			want:          false,
		},
		{
			name:          "failed to cancelled",
			currentStatus: "failed",
			nextStatus:    "cancelled",
			want:          false,
		},
		{
			name:          "failed to delivered",
			currentStatus: "failed",
			nextStatus:    "delivered",
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
			nextStatus:    "packed",
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
			nextStatus:    "packed",
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
			got := isValidShippingTransition(tt.currentStatus, tt.nextStatus)

			if got != tt.want {
				t.Errorf(
					"isValidShippingTransition(%q, %q) = %v; want %v",
					tt.currentStatus,
					tt.nextStatus,
					got,
					tt.want,
				)
			}
		})
	}
}

type shipmentFixture struct {
	uc                 *ShipmentUseCase
	shipmentRepo       *mocks.MockShipmentRepository
	orderRepo          *mocks.MockOrderRepository
	uow                *mocks.MockUnitOfWork
	txShipment         *mocks.MockShipmentRepository
	txOrder            *mocks.MockOrderRepository
	txHistory          *mocks.MockOrderStatusHistoryRepository
	txReturn           *mocks.MockReturnRequestRepository
	txReferral         *mocks.MockReferralRepository
	txWallet           *mocks.MockWalletRepository
	txUser             *mocks.MockUserRepository
	txNotification     *mocks.MockNotificationRepository
	txOrderItem        *mocks.MockOrderItemRepository
	txOrderAddress     *mocks.MockOrderAddressRepository
	txReferralSettings *mocks.MockReferralSettingsRepository
	txPayment          *mocks.MockPaymentRepository
	txCoupon           *mocks.MockCouponRepository
	txCouponUsage      *mocks.MockCouponUsageRepository
	txWalletTopUp      *mocks.MockWalletTopUpRepository
	txAddress          *mocks.MockAddressRepository
	txProductVariant   *mocks.MockProductVariantRepository
	txCart             *mocks.MockCartRepository
}

func newShipmentFixture(t *testing.T) *shipmentFixture {
	t.Helper()
	shipmentRepo := mocks.NewMockShipmentRepository(t)
	orderRepo := mocks.NewMockOrderRepository(t)
	uow := mocks.NewMockUnitOfWork(t)
	txShipment := mocks.NewMockShipmentRepository(t)
	txOrder := mocks.NewMockOrderRepository(t)
	txHistory := mocks.NewMockOrderStatusHistoryRepository(t)
	txReturn := mocks.NewMockReturnRequestRepository(t)
	txReferral := mocks.NewMockReferralRepository(t)
	txWallet := mocks.NewMockWalletRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &shipmentFixture{
		uc:                 NewShipmentUseCase(shipmentRepo, orderRepo, uow, log),
		shipmentRepo:       shipmentRepo,
		orderRepo:          orderRepo,
		uow:                uow,
		txShipment:         txShipment,
		txOrder:            txOrder,
		txHistory:          txHistory,
		txReturn:           txReturn,
		txReferral:         txReferral,
		txWallet:           txWallet,
		txUser:             mocks.NewMockUserRepository(t),
		txNotification:     mocks.NewMockNotificationRepository(t),
		txOrderItem:        mocks.NewMockOrderItemRepository(t),
		txOrderAddress:     mocks.NewMockOrderAddressRepository(t),
		txReferralSettings: mocks.NewMockReferralSettingsRepository(t),
		txPayment:          mocks.NewMockPaymentRepository(t),
		txCoupon:           mocks.NewMockCouponRepository(t),
		txCouponUsage:      mocks.NewMockCouponUsageRepository(t),
		txWalletTopUp:      mocks.NewMockWalletTopUpRepository(t),
		txAddress:          mocks.NewMockAddressRepository(t),
		txProductVariant:   mocks.NewMockProductVariantRepository(t),
		txCart:             mocks.NewMockCartRepository(t),
	}
}

func (f *shipmentFixture) expectTX(ctx context.Context) {
	f.uow.EXPECT().WithinTransaction(ctx, mock.AnythingOfType("func(interfaces.TransactionRepositories) error")).
		RunAndReturn(func(_ context.Context, fn func(interfaces.TransactionRepositories) error) error {
			return fn(interfaces.TransactionRepositories{
				User:               f.txUser,
				Notification:       f.txNotification,
				Order:              f.txOrder,
				OrderItem:          f.txOrderItem,
				OrderAddress:       f.txOrderAddress,
				OrderStatusHistory: f.txHistory,
				Shipment:           f.txShipment,
				ReturnRequest:      f.txReturn,
				Referral:           f.txReferral,
				ReferralSettings:   f.txReferralSettings,
				Wallet:             f.txWallet,
				Payment:            f.txPayment,
				Coupon:             f.txCoupon,
				CouponUsage:        f.txCouponUsage,
				WalletTopUp:        f.txWalletTopUp,
				Address:            f.txAddress,
				ProductVariant:     f.txProductVariant,
				Cart:               f.txCart,
			})
		}).Once()
}

func TestShipmentUseCase_CRUDAndQueries(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()

	t.Run("CreateShipment", func(t *testing.T) {
		f := newShipmentFixture(t)
		assert.ErrorIs(t, f.uc.CreateShipment(ctx, nil), domainerrors.ErrInvalidShipment)
		f.orderRepo.EXPECT().GetByID(ctx, uuid.Nil).Return(nil, domainerrors.ErrRecordNotFound).Once()
		assert.ErrorIs(t, f.uc.CreateShipment(ctx, &models.Shipment{}), domainerrors.ErrRecordNotFound)

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{
			ID:          orderID,
			OrderStatus: "pending",
		}, nil).Once()
		assert.ErrorIs(t, f.uc.CreateShipment(ctx, &models.Shipment{OrderID: orderID}), domainerrors.ErrOrderNeedToConfirm)

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{
			ID:          orderID,
			OrderStatus: "confirmed",
		}, nil).Once()
		f.shipmentRepo.EXPECT().GetByOrderID(ctx, orderID).Return(nil, domainerrors.ErrRecordNotFound).Maybe()
		f.shipmentRepo.EXPECT().Create(ctx, mock.MatchedBy(func(s *models.Shipment) bool {
			return s.OrderID == orderID
		})).Return(nil).Once()

		require.NoError(t, f.uc.CreateShipment(ctx, &models.Shipment{OrderID: orderID}))
	})

	t.Run("GetShipmentByID GetShipmentByOrderID GetTrackingHistory", func(t *testing.T) {
		f := newShipmentFixture(t)
		_, err := f.uc.GetShipmentByID(ctx, 0)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidShipmentID)

		ship := &models.Shipment{ID: 3, OrderID: orderID}
		f.shipmentRepo.EXPECT().GetByID(ctx, uint(3)).Return(ship, nil).Once()
		got, err := f.uc.GetShipmentByID(ctx, 3)
		require.NoError(t, err)
		assert.Equal(t, ship, got)

		_, err = f.uc.GetShipmentByOrderID(ctx, uuid.Nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidOrderID)

		f.shipmentRepo.EXPECT().GetByOrderID(ctx, orderID).Return(ship, nil).Once()
		got, err = f.uc.GetShipmentByOrderID(ctx, orderID)
		require.NoError(t, err)
		assert.Equal(t, ship, got)

		_, err = f.uc.GetTrackingHistory(ctx, 0)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidShipmentID)

		hist := []models.ShipmentTrackingHistory{{ID: 1, ShipmentID: 3}}
		f.shipmentRepo.EXPECT().GetByID(ctx, uint(3)).Return(ship, nil).Maybe()
		f.shipmentRepo.EXPECT().GetTrackingHistory(ctx, uint(3)).Return(hist, nil).Once()
		gotHist, err := f.uc.GetTrackingHistory(ctx, 3)
		require.NoError(t, err)
		assert.Equal(t, hist, gotHist)
	})

	t.Run("UpdateShipment", func(t *testing.T) {
		f := newShipmentFixture(t)
		assert.ErrorIs(t, f.uc.UpdateShipment(ctx, nil), domainerrors.ErrInvalidShipment)
		assert.ErrorIs(t, f.uc.UpdateShipment(ctx, &models.Shipment{}), domainerrors.ErrInvalidShipmentID)

		courier := "BlueDart"
		existing := &models.Shipment{ID: 4, OrderID: orderID, ShippingStatus: "pending"}
		f.shipmentRepo.EXPECT().GetByID(ctx, uint(4)).Return(existing, nil).Once()
		f.shipmentRepo.EXPECT().Update(ctx, mock.MatchedBy(func(s *models.Shipment) bool {
			return s.CourierName != nil && *s.CourierName == "BlueDart"
		})).Return(nil).Once()

		require.NoError(t, f.uc.UpdateShipment(ctx, &models.Shipment{ID: 4, CourierName: &courier}))
	})

	t.Run("GetCustomerShipment and GetCustomerTrackingHistory", func(t *testing.T) {
		f := newShipmentFixture(t)
		_, err := f.uc.GetCustomerShipment(ctx, 0, orderID)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidUserID)
		_, err = f.uc.GetCustomerShipment(ctx, 1, uuid.Nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidOrderID)

		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 99}, nil).Once()
		_, err = f.uc.GetCustomerShipment(ctx, 1, orderID)
		assert.ErrorIs(t, err, domainerrors.ErrRecordNotFound)

		ship := &models.Shipment{ID: 8, OrderID: orderID}
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 1}, nil).Once()
		f.shipmentRepo.EXPECT().GetByOrderID(ctx, orderID).Return(ship, nil).Once()
		got, err := f.uc.GetCustomerShipment(ctx, 1, orderID)
		require.NoError(t, err)
		assert.Equal(t, ship, got)

		_, err = f.uc.GetCustomerTrackingHistory(ctx, 0, orderID)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidUserID)

		hist := []models.ShipmentTrackingHistory{{ID: 10, ShipmentID: 8}}
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 1}, nil).Once()
		f.shipmentRepo.EXPECT().GetByOrderID(ctx, orderID).Return(ship, nil).Once()
		f.shipmentRepo.EXPECT().GetTrackingHistory(ctx, uint(8)).Return(hist, nil).Once()
		gotHist, err := f.uc.GetCustomerTrackingHistory(ctx, 1, orderID)
		require.NoError(t, err)
		assert.Equal(t, hist, gotHist)
	})
}

func TestShipmentUseCase_UpdateShippingStatus(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()

	t.Run("validation and invalid transition", func(t *testing.T) {
		f := newShipmentFixture(t)
		assert.ErrorIs(t, f.uc.UpdateShippingStatus(ctx, 0, "shipped", nil, nil), domainerrors.ErrInvalidShipmentID)
		assert.ErrorIs(t, f.uc.UpdateShippingStatus(ctx, 1, "bad_status", nil, nil), domainerrors.ErrInvalidShippingStatus)

		f.expectTX(ctx)
		f.txShipment.EXPECT().GetByID(ctx, uint(5)).Return(&models.Shipment{
			ID:             5,
			OrderID:        orderID,
			Type:           models.ShipmentTypeDelivery,
			ShippingStatus: "delivered",
		}, nil).Once()
		err := f.uc.UpdateShippingStatus(ctx, 5, "cancelled", nil, nil)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidShippingTransition)
	})

	t.Run("delivery shipped to out_for_delivery", func(t *testing.T) {
		f := newShipmentFixture(t)
		desc := "Left hub"
		loc := "Kochi"
		ship := &models.Shipment{ID: 5, OrderID: orderID, Type: models.ShipmentTypeDelivery, ShippingStatus: "in_transit"}

		f.shipmentRepo.EXPECT().GetByID(ctx, uint(5)).Return(ship, nil).Maybe()
		f.expectTX(ctx)
		f.txShipment.EXPECT().GetByID(ctx, uint(5)).Return(ship, nil).Maybe()
		f.txShipment.EXPECT().UpdateShippingStatus(ctx, uint(5), "out_for_delivery", &desc, &loc).Return(nil).Once()

		require.NoError(t, f.uc.UpdateShippingStatus(ctx, 5, "out_for_delivery", &desc, &loc))
	})

	t.Run("delivery out_for_delivery to delivered marks order delivered", func(t *testing.T) {
		f := newShipmentFixture(t)
		ship := &models.Shipment{ID: 6, OrderID: orderID, Type: models.ShipmentTypeDelivery, ShippingStatus: "out_for_delivery"}
		ord := &models.Order{ID: orderID, UserID: 2, OrderStatus: "shipped", PaymentStatus: "paid", TotalAmount: 500}

		f.shipmentRepo.EXPECT().GetByID(ctx, uint(6)).Return(ship, nil).Maybe()
		f.expectTX(ctx)
		f.txShipment.EXPECT().GetByID(ctx, uint(6)).Return(ship, nil).Maybe()
		f.txShipment.EXPECT().UpdateShippingStatus(ctx, uint(6), "delivered", (*string)(nil), (*string)(nil)).Return(nil).Once()
		f.txOrder.EXPECT().LockByID(ctx, orderID).Return(ord, nil).Maybe()
		f.txOrder.EXPECT().GetByID(ctx, orderID).Return(ord, nil).Maybe()
		f.txOrder.EXPECT().Update(ctx, mock.AnythingOfType("*models.Order")).Return(nil).Maybe()
		f.txOrder.EXPECT().UpdateOrderStatus(ctx, orderID, "delivered", mock.Anything).Return(nil).Maybe()
		f.txHistory.EXPECT().Create(ctx, mock.AnythingOfType("*models.OrderStatusHistory")).Return(nil).Maybe()
		f.txReferral.EXPECT().GetByReferrerID(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, int64(0), nil).Maybe()
		f.txReferral.EXPECT().GetReferrals(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, int64(0), nil).Maybe()

		require.NoError(t, f.uc.UpdateShippingStatus(ctx, 6, "delivered", nil, nil))
	})

	t.Run("return shipment updates return request status", func(t *testing.T) {
		f := newShipmentFixture(t)
		retID := uint(12)
		ship := &models.Shipment{
			ID:              7,
			OrderID:         orderID,
			ReturnRequestID: &retID,
			Type:            models.ShipmentTypeReturn,
			ShippingStatus:  "out_for_delivery",
		}
		retReq := &models.ReturnRequest{ID: 12, OrderID: orderID, Status: ReturnStatusInTransit}

		f.expectTX(ctx)
		f.txShipment.EXPECT().GetByID(ctx, uint(7)).Return(ship, nil).Once()
		f.txShipment.EXPECT().UpdateShippingStatus(ctx, uint(7), "delivered", (*string)(nil), (*string)(nil)).Return(nil).Once()
		f.txReturn.EXPECT().GetByID(ctx, retID).Return(retReq, nil).Once()
		f.txReturn.EXPECT().Update(ctx, mock.AnythingOfType("*models.ReturnRequest")).Return(nil).Once()

		require.NoError(t, f.uc.UpdateShippingStatus(ctx, 7, "delivered", nil, nil))
	})

	t.Run("return_to_customer delivered closes return request", func(t *testing.T) {
		f := newShipmentFixture(t)
		retID := uint(15)
		ship := &models.Shipment{
			ID:              8,
			OrderID:         orderID,
			ReturnRequestID: &retID,
			Type:            models.ShipmentTypeReturnToCustomer,
			ShippingStatus:  "out_for_delivery",
		}
		retReq := &models.ReturnRequest{ID: 15, OrderID: orderID, Status: ReturnStatusInspectionFailed}

		f.shipmentRepo.EXPECT().GetByID(ctx, uint(8)).Return(ship, nil).Maybe()
		f.expectTX(ctx)
		f.txShipment.EXPECT().GetByID(ctx, uint(8)).Return(ship, nil).Maybe()
		f.txShipment.EXPECT().UpdateShippingStatus(ctx, uint(8), "delivered", (*string)(nil), (*string)(nil)).Return(nil).Once()
		f.txReturn.EXPECT().LockByID(ctx, retID).Return(retReq, nil).Maybe()
		f.txReturn.EXPECT().GetByID(ctx, retID).Return(retReq, nil).Maybe()
		f.txReturn.EXPECT().Update(ctx, mock.MatchedBy(func(r *models.ReturnRequest) bool {
			return r.Status == ReturnStatusClosed
		})).Return(nil).Once()

		require.NoError(t, f.uc.UpdateShippingStatus(ctx, 8, "delivered", nil, nil))
	})

	t.Run("tx error propagates", func(t *testing.T) {
		f := newShipmentFixture(t)
		txErr := errors.New("tx failed")
		f.shipmentRepo.EXPECT().GetByID(ctx, uint(5)).Return(&models.Shipment{
			ID: 5, OrderID: orderID, Type: models.ShipmentTypeDelivery, ShippingStatus: "pending",
		}, nil).Maybe()
		f.uow.EXPECT().WithinTransaction(ctx, mock.Anything).Return(txErr).Once()

		assert.ErrorIs(t, f.uc.UpdateShippingStatus(ctx, 5, "packed", nil, nil), txErr)
	})
}
