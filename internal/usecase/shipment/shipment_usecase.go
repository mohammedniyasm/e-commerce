package shipment

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	// "uuid"
)

const (
	RequestTypeReturn    = "return"
	ReturnStatusReturn   = "return"
	ReturnStatusExchange = "exchange"
	RequestTypeExchange  = "exchange"

	ReturnStatusRequested         = "requested"
	ReturnStatusApproved          = "approved"
	ReturnStatusRejected          = "rejected"
	ReturnStatusPickupScheduled   = "pickup_scheduled"
	ReturnStatusInTransit         = "in_transit"
	ReturnStatusReceived          = "received"
	ReturnStatusInspectionPending = "inspection_pending"
	ReturnStatusInspectionPassed  = "inspection_passed"
	ReturnStatusInspectionFailed  = "inspection_failed"
	ReturnStatusClosed            = "closed"
)

type ShipmentUseCase struct {
	shipmentRepo interfaces.ShipmentRepository
	orderRepo    interfaces.OrderRepository
	unitOfWork   interfaces.UnitOfWork
	log          *slog.Logger
}

func NewShipmentUseCase(shipmentRepo interfaces.ShipmentRepository, orderRepo interfaces.OrderRepository, unitOfWork interfaces.UnitOfWork, log *slog.Logger) *ShipmentUseCase {
	return &ShipmentUseCase{
		shipmentRepo: shipmentRepo,
		orderRepo:    orderRepo,
		unitOfWork:   unitOfWork,
		log:          log,
	}
}
func isValidShippingTransition(currentStatus, nextStatus string) bool {
	allowedTransitions := map[string][]string{
		"pending": {
			"packed",
			"cancelled",
		},
		"packed": {
			"ready_to_ship",
			"cancelled",
		},
		"ready_to_ship": {
			"in_transit",
			"cancelled",
		},
		"in_transit": {
			"out_for_delivery",
			"failed",
			"returned",
		},
		"out_for_delivery": {
			"delivered",
			"failed",
			"returned",
		},
		"failed": {
			"in_transit",
			"out_for_delivery",
			"returned",
		},
	}
	for _, allowedStatus := range allowedTransitions[currentStatus] {
		if allowedStatus == nextStatus {
			return true
		}
	}
	return false
}
func (u *ShipmentUseCase) CreateShipment(ctx context.Context, shipment *models.Shipment) error {
	if shipment == nil {
		return domainerrors.ErrInvalidShipment
	}
	order, err := u.orderRepo.GetByID(ctx, shipment.OrderID)
	if err != nil {
		return err
	}
	if order.OrderStatus == "confirmed" || order.OrderStatus == "processing" {
		shipment.ShippingStatus = "pending"
		shipment.Type = models.ShipmentTypeDelivery
		if err := u.shipmentRepo.Create(ctx, shipment); err != nil {
			return err
		}
		return nil
	}
	return domainerrors.ErrOrderNeedToConfirm
}
func (u *ShipmentUseCase) GetShipmentByID(ctx context.Context, id uint) (*models.Shipment, error) {
	if id <= 0 {
		return nil, domainerrors.ErrInvalidShipmentID
	}
	shipment, err := u.shipmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return shipment, nil
}
func (u *ShipmentUseCase) GetShipmentByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Shipment, error) {
	if orderID == uuid.Nil {
		return nil, domainerrors.ErrInvalidOrderID
	}
	return u.shipmentRepo.GetByOrderID(ctx, orderID)
}
func (u *ShipmentUseCase) UpdateShipment(ctx context.Context, shipment *models.Shipment) error {
	if shipment == nil {
		return domainerrors.ErrInvalidShipment
	}
	if shipment.ID == 0 {
		return domainerrors.ErrInvalidShipmentID
	}
	if _, err := u.shipmentRepo.GetByID(ctx, shipment.ID); err != nil {
		return err
	}
	return u.shipmentRepo.Update(ctx, shipment)
}
func (u *ShipmentUseCase) UpdateShippingStatus(ctx context.Context, shipmentID uint, status string, description *string, location *string) error {
	if shipmentID == 0 {
		return domainerrors.ErrInvalidShipmentID
	}
	validStatuses := map[string]bool{
		"pending":          true,
		"packed":           true,
		"ready_to_ship":    true,
		"in_transit":       true,
		"out_for_delivery": true,
		"delivered":        true,
		"failed":           true,
		"returned":         true,
		"cancelled":        true,
	}
	if !validStatuses[status] {
		return domainerrors.ErrInvalidShippingStatus
	}
	err := u.unitOfWork.WithinTransaction(ctx, func(repos interfaces.TransactionRepositories) error {
		shipment, err := repos.Shipment.GetByID(ctx, shipmentID)
		if err != nil {
			return err
		}
		if !isValidShippingTransition(shipment.ShippingStatus, status) {
			return domainerrors.ErrInvalidShippingTransition
		}
		if err := repos.Shipment.UpdateShippingStatus(ctx, shipmentID, status, description, location); err != nil {
			return err
		}
		if shipment.Type == models.ShipmentTypeReturnToCustomer {
			if status != "delivered" {
				return nil
			}
			if shipment.ReturnRequestID == nil {
				return domainerrors.ErrInvalidInput
			}
			request, err := repos.ReturnRequest.GetByID(
				ctx,
				*shipment.ReturnRequestID,
			)
			if err != nil {
				return fmt.Errorf("get return request: %w", err)
			}
			if request.Status != ReturnStatusInspectionFailed {
				return domainerrors.ErrInvalidOrderTransition
			}
			if !isValidReturnTransition(
				request.Status,
				ReturnStatusClosed,
			) {
				return domainerrors.ErrInvalidOrderTransition
			}
			request.Status = ReturnStatusClosed
			if err := repos.ReturnRequest.Update(ctx, request); err != nil {
				return fmt.Errorf("close return request: %w", err)
			}
			return nil
		}
		if shipment.Type == models.ShipmentTypeReturn {
			if status == ReturnStatusInTransit || status == "delivered" {
				if shipment.ReturnRequestID == nil {
					return domainerrors.ErrInvalidInput
				}
				request, err := repos.ReturnRequest.GetByID(
					ctx,
					*shipment.ReturnRequestID,
				)
				if err != nil {
					return fmt.Errorf("get return request: %w", err)
				}
				var nextReturnStatus string
				switch status {
				case "in_transit":
					nextReturnStatus = ReturnStatusInTransit

				case "delivered":
					nextReturnStatus = ReturnStatusReceived
				}

				if !isValidReturnTransition(
					request.Status,
					nextReturnStatus,
				) {
					return domainerrors.ErrInvalidOrderTransition
				}
				request.Status = nextReturnStatus
				if err := repos.ReturnRequest.Update(ctx, request); err != nil {
					return fmt.Errorf("update return request status: %w", err)
				}
			}
			return nil
		}
		var orderStatus string
		switch status {
		case "shipped":
			orderStatus = "shipped"
		case "in_transit":
			orderStatus = "shipped"
		case "delivered":
			orderStatus = "delivered"
		}
		if orderStatus != "" {
			note := "Order status automatically updated to " +
				orderStatus + " because the shipment status changed."
			if err := repos.Order.UpdateOrderStatus(ctx, shipment.OrderID, orderStatus, &note); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to update shipping status",
			"shipment_id", shipmentID,
			"status", status,
			"error", err,
		)
		return err
	}
	u.log.InfoContext(
		ctx,
		"shipping status updated successfully",
		"shipment_id", shipmentID,
		"status", status,
	)
	return nil
}
func (u *ShipmentUseCase) GetTrackingHistory(ctx context.Context, shipmentID uint) ([]models.ShipmentTrackingHistory, error) {
	if shipmentID == 0 {
		return nil, domainerrors.ErrInvalidShipmentID
	}
	_, err := u.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	return u.shipmentRepo.GetTrackingHistory(ctx, shipmentID)
}
func (u *ShipmentUseCase) GetCustomerShipment(ctx context.Context, userID uint, orderID uuid.UUID) (*models.Shipment, error) {
	if userID == 0 {
		return nil, domainerrors.ErrInvalidUserID
	}
	if orderID == uuid.Nil {
		return nil, domainerrors.ErrInvalidOrderID
	}
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, domainerrors.ErrRecordNotFound
	}
	shipment, err := u.shipmentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return shipment, nil
}
func (u *ShipmentUseCase) GetCustomerTrackingHistory(ctx context.Context, userID uint, orderID uuid.UUID) ([]models.ShipmentTrackingHistory, error) {
	shipment, err := u.GetCustomerShipment(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	history, err := u.shipmentRepo.GetTrackingHistory(
		ctx,
		shipment.ID,
	)
	if err != nil {
		return nil, err
	}
	return history, nil
}
func isValidReturnStatus(status string) bool {
	switch status {
	case ReturnStatusRequested,
		ReturnStatusApproved,
		ReturnStatusRejected,
		ReturnStatusPickupScheduled,
		ReturnStatusInTransit,
		ReturnStatusReceived,
		ReturnStatusInspectionPending,
		ReturnStatusInspectionPassed,
		ReturnStatusInspectionFailed,
		ReturnStatusClosed:
		return true
	default:
		return false
	}
}
func isValidReturnTransition(current, next string) bool {
	if current == next {
		return false
	}

	switch current {
	case ReturnStatusRequested:
		return next == ReturnStatusApproved ||
			next == ReturnStatusRejected

	case ReturnStatusApproved:
		return next == ReturnStatusPickupScheduled ||
			next == ReturnStatusClosed

	case ReturnStatusPickupScheduled:
		return next == ReturnStatusInTransit ||
			next == ReturnStatusClosed

	case ReturnStatusInTransit:
		return next == ReturnStatusReceived

	case ReturnStatusReceived:
		return next == ReturnStatusInspectionPending

	case ReturnStatusInspectionPending:
		return next == ReturnStatusInspectionPassed ||
			next == ReturnStatusInspectionFailed

	case ReturnStatusInspectionPassed:
		return next == ReturnStatusClosed

	case ReturnStatusInspectionFailed:
		return next == ReturnStatusClosed

	default:
		return false
	}
}

var _ interfaces.ShipmentUseCase = (*ShipmentUseCase)(nil)
