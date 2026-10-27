package return_request

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
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

type ReturnRequestUseCase struct {
	returnRequestRepo interfaces.ReturnRequestRepository
	orderRepo         interfaces.OrderRepository
	orderItemRepo     interfaces.OrderItemRepository
	variantRepo       interfaces.ProductVariantRepository
	uow               interfaces.UnitOfWork
	log               *slog.Logger
}

func NewReturnRequestUseCase(
	returnRequestRepo interfaces.ReturnRequestRepository,
	orderRepo interfaces.OrderRepository,
	orderItemRepo interfaces.OrderItemRepository,
	variantRepo interfaces.ProductVariantRepository,
	uow interfaces.UnitOfWork,
	log *slog.Logger,
) *ReturnRequestUseCase {
	return &ReturnRequestUseCase{
		returnRequestRepo: returnRequestRepo,
		orderRepo:         orderRepo,
		orderItemRepo:     orderItemRepo,
		variantRepo:       variantRepo,
		uow:               uow,
		log:               log,
	}
}

func (u *ReturnRequestUseCase) Create(ctx context.Context, userID uint, orderID uuid.UUID, orderItemID uint, requestType string, requestedVariantID *uint, quantity int, reason string) (*models.ReturnRequest, error) {
	if userID == 0 ||
		orderID == uuid.Nil ||
		orderItemID == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	requestType = strings.ToLower(strings.TrimSpace(requestType))
	reason = strings.TrimSpace(reason)
	if requestType != ReturnStatusReturn &&
		requestType != ReturnStatusExchange {
		return nil, domainerrors.ErrInvalidInput
	}
	if quantity <= 0 {
		return nil, domainerrors.ErrInvalidQuantity
	}
	if reason == "" {
		return nil, domainerrors.ErrInvalidInput
	}
	if requestType == ReturnStatusReturn &&
		requestedVariantID != nil {
		return nil, domainerrors.ErrInvalidInput
	}
	if requestType == ReturnStatusExchange &&
		(requestedVariantID == nil || *requestedVariantID == 0) {
		return nil, domainerrors.ErrInvalidInput
	}
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to retrieve order for return request",
			"order_id", orderID,
			"error", err,
		)
		return nil, err
	}
	if order.UserID != userID {
		return nil, domainerrors.ErrInvalidInput
	}
	if order.OrderStatus != "delivered" {
		return nil, domainerrors.ErrInvalidOrderStatus
	}
	orderItem, err := u.orderItemRepo.GetByID(ctx, orderItemID)
	if err != nil {
		return nil, err
	}
	if orderItem.OrderID != order.ID {
		return nil, domainerrors.ErrInvalidInput
	}
	if quantity > orderItem.Quantity {
		return nil, domainerrors.ErrInvalidQuantity
	}
	if requestType == ReturnStatusExchange {
		variant, err := u.variantRepo.GetByID(
			ctx,
			int64(*requestedVariantID),
		)
		if err != nil {
			return nil, err
		}

		if !variant.IsActive ||
			variant.ID == orderItem.VariantID {
			return nil, domainerrors.ErrProductNotAvailable
		}
	}

	request := &models.ReturnRequest{
		OrderID:            order.ID,
		OrderItemID:        orderItemID,
		RequestType:        requestType,
		RequestedVariantID: requestedVariantID,
		Quantity:           quantity,
		Reason:             reason,
		Status:             ReturnStatusRequested,
	}

	err = u.uow.WithinTransaction(
		ctx,
		func(repos interfaces.TransactionRepositories) error {
			hasActiveRequest, err := repos.ReturnRequest.HasActiveRequest(
				ctx,
				orderItemID,
			)
			if err != nil {
				return err
			}

			if hasActiveRequest {
				return domainerrors.ErrInvalidInput
			}

			if err := repos.ReturnRequest.Create(ctx, request); err != nil {
				u.log.ErrorContext(
					ctx,
					"failed to create return request",
					"order_id", order.ID,
					"order_item_id", orderItemID,
					"error", err,
				)
				return err
			}

			adminIDs, err := repos.User.ListAdminIDs(ctx)
			if err != nil {
				u.log.ErrorContext(
					ctx,
					"failed to retrieve admin IDs for return request notification",
					"return_request_id", request.ID,
					"error", err,
				)
				return err
			}

			title := "New " + requestType + " request"
			message := "A customer has submitted a " +
				requestType + " request for order " +
				order.OrderNumber + "."

			entityType := "return_request"
			entityID := strconv.FormatUint(uint64(request.ID), 10)

			eventData, err := json.Marshal(map[string]interface{}{
				"return_request_id": request.ID,
				"order_id":          request.OrderID,
				"order_item_id":     request.OrderItemID,
				"request_type":      request.RequestType,
				"status":            request.Status,
			})
			if err != nil {
				return err
			}

			rawEventData := json.RawMessage(eventData)

			notification := &models.Notification{
				ID:         uuid.New(),
				Type:       "return_request_created",
				Title:      title,
				Message:    message,
				Priority:   "medium",
				EntityType: &entityType,
				EntityID:   &entityID,
				EventData:  &rawEventData,
				CreatedAt:  time.Now(),
			}

			if err := repos.Notification.Create(
				ctx,
				notification,
			); err != nil {
				u.log.ErrorContext(
					ctx,
					"failed to create return request notification",
					"return_request_id", request.ID,
					"error", err,
				)
				return err
			}

			recipients := make(
				[]models.NotificationRecipient,
				0,
				len(adminIDs),
			)

			for _, adminID := range adminIDs {
				recipients = append(
					recipients,
					models.NotificationRecipient{
						NotificationID: notification.ID,
						AdminID:        adminID,
					},
				)
			}

			if len(recipients) > 0 {
				if err := repos.Notification.CreateRecipients(
					ctx,
					recipients,
				); err != nil {
					u.log.ErrorContext(
						ctx,
						"failed to create return request notification recipients",
						"notification_id", notification.ID,
						"error", err,
					)
					return err
				}
			}

			u.log.InfoContext(
				ctx,
				"return request created successfully",
				"return_request_id", request.ID,
				"order_id", request.OrderID,
				"request_type", request.RequestType,
				"notification_id", notification.ID,
				"admin_recipient_count", len(recipients),
			)

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return request, nil
}
func (u *ReturnRequestUseCase) GetByID(ctx context.Context, requestID uint, userID uint) (*models.ReturnRequest, error) {
	if requestID == 0 || userID == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	request, err := u.returnRequestRepo.GetByIDAndUserID(ctx, requestID, userID)
	if err != nil {
		return nil, fmt.Errorf("get customer return request: %w", err)
	}
	return request, nil
}
func (u *ReturnRequestUseCase) ListByUserID(ctx context.Context, userID uint, limit int, page int) ([]models.ReturnRequest, int64, error) {
	if userID == 0 {
		return nil, 0, domainerrors.ErrInvalidInput
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}
	requests, total, err := u.returnRequestRepo.ListByUserID(
		ctx,
		userID,
		limit,
		page,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list customer return requests: %w", err)
	}
	return requests, total, nil
}
func (u *ReturnRequestUseCase) List(ctx context.Context, orderID *uuid.UUID, status *string, limit int, page int) ([]models.ReturnRequest, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}
	if orderID != nil && *orderID == uuid.Nil {
		return nil, 0, domainerrors.ErrInvalidInput
	}
	if status != nil {
		normalized := strings.ToLower(strings.TrimSpace(*status))
		if normalized != "" && !isValidReturnStatus(normalized) {
			return nil, 0, domainerrors.ErrInvalidInput
		}
		status = &normalized
	}
	requests, total, err := u.returnRequestRepo.List(
		ctx,
		orderID,
		status,
		limit,
		page,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list return requests: %w", err)
	}
	return requests, total, nil
}
func (u *ReturnRequestUseCase) UpdateStatus(ctx context.Context, returnRequestID uint, status string, adminNote *string) (*models.ReturnRequest, error) {
	if returnRequestID == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if !isValidReturnStatus(status) {
		return nil, domainerrors.ErrInvalidInput
	}
	var updatedRequest *models.ReturnRequest
	err := u.uow.WithinTransaction(
		ctx,
		func(repos interfaces.TransactionRepositories) error {
			request, err := repos.ReturnRequest.GetByID(
				ctx,
				returnRequestID,
			)
			if err != nil {
				return fmt.Errorf(
					"get return request for status update: %w",
					err,
				)
			}
			if !isValidReturnTransition(request.Status, status) {
				return domainerrors.ErrInvalidOrderTransition
			}
			request.Status = status
			request.AdminNote = adminNote
			if err := repos.ReturnRequest.Update(
				ctx,
				request,
			); err != nil {
				return fmt.Errorf(
					"update return request status: %w",
					err,
				)
			}
			updatedRequest = request
			return nil
		},
	)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to update return request status",
			"return_request_id", returnRequestID,
			"status", status,
			"error", err,
		)
		return nil, err
	}
	u.log.InfoContext(
		ctx,
		"return request status updated",
		"return_request_id", returnRequestID,
		"status", status,
	)
	return updatedRequest, nil
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

var _ interfaces.ReturnRequestUseCase = (*ReturnRequestUseCase)(nil)
