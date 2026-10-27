package order

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	// "uuid"
)

type OrderUseCase struct {
	orderRepo        interfaces.OrderRepository
	unitOfWork       interfaces.UnitOfWork
	invoiceGenerator interfaces.InvoiceGenerator
	log              *slog.Logger
}

func NewOrderUseCase(
	orderRepo interfaces.OrderRepository,
	unitOfWork interfaces.UnitOfWork,
	invoiceGenerator interfaces.InvoiceGenerator,
	log *slog.Logger,
) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:        orderRepo,
		unitOfWork:       unitOfWork,
		invoiceGenerator: invoiceGenerator,
		log:              log,
	}
}
func isValidOrderTransition(currentStatus, nextStatus string) bool {
	allowedTransitions := map[string][]string{
		"pending": {
			"confirmed",
			"cancelled",
		},
		"confirmed": {
			"processing",
			"cancelled",
		},
		"processing": {
			"shipped",
			"cancelled",
		},
		"shipped": {
			"delivered",
			"returned",
		},
		"out_for_delivery": {
			"delivered",
			"cancelled",
		},
		"delivered": {
			"returned",
			"processing",
		},
	}

	for _, allowedStatus := range allowedTransitions[currentStatus] {
		if allowedStatus == nextStatus {
			return true
		}
	}

	return false
}
func (u *OrderUseCase) GetByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error) {
	if orderID == uuid.Nil {
		return nil, domainerrors.ErrInvalidOrderID
	}
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return order, nil
}
func (u *OrderUseCase) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	if orderNumber == "" {
		return nil, domainerrors.ErrInvalidOrderNumber
	}
	return u.orderRepo.GetByOrderNumber(ctx, orderNumber)
}
func (u *OrderUseCase) ListByUserID(ctx context.Context, userID uint, limit int, page int, sort *string) ([]models.Order, int64, error) {
	if userID == 0 {
		return nil, 0, domainerrors.ErrInvalidUserID
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
	orders, total, err := u.orderRepo.ListByUserID(ctx, userID, limit, page, sort)
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
func (u *OrderUseCase) List(ctx context.Context, userID *uint, search string, limit int, page int, sort *string, orderStatus *string, paymentStatus *string) ([]models.Order, int64, error) {
	if userID != nil {
		if *userID <= 0 {
			return nil, 0, domainerrors.ErrInvalidUserID
		}
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
	orders, total, err := u.orderRepo.List(ctx, userID, search, limit, page, sort, orderStatus, paymentStatus)
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
func (u *OrderUseCase) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, note *string) error {
	if orderID == uuid.Nil {
		return domainerrors.ErrInvalidOrderID
	}
	validStatuses := map[string]bool{
		"pending":          true,
		"confirmed":        true,
		"processing":       true,
		"shipped":          true,
		"out_for_delivery": true,
		"delivered":        true,
		"cancelled":        true,
		"returned":         true,
	}
	if !validStatuses[status] {
		return domainerrors.ErrInvalidOrderStatus
	}
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if !isValidOrderTransition(order.OrderStatus, status) {
		return domainerrors.ErrInvalidOrderTransition
	}
	return u.orderRepo.UpdateOrderStatus(ctx, orderID, status, note)
}
func (u *OrderUseCase) PlaceOrder(ctx context.Context, userID uint, cartItemIDs []uint, addressID uint, paymentMethod string, couponCode *string) (*models.Order, error) {
	if userID <= 0 {
		return nil, domainerrors.ErrInvalidUserID
	}
	if addressID <= 0 {
		return nil, domainerrors.ErrInvalidAddress
	}
	if len(cartItemIDs) == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	paymentMethod = strings.ToLower(strings.TrimSpace(paymentMethod))
	switch paymentMethod {
	case "cod", "online":
		// Supported payment method.

	default:
		return nil, domainerrors.ErrInvalidInput
	}
	if couponCode != nil {
		normalizedCode := strings.TrimSpace(*couponCode)
		if normalizedCode == "" {
			couponCode = nil
		} else {
			couponCode = &normalizedCode
		}
	}
	selectedIDs := make(map[uint]struct{})
	for _, itemID := range cartItemIDs {
		if itemID <= 0 {
			return nil, domainerrors.ErrInvalidInput
		}
		if _, exists := selectedIDs[itemID]; exists {
			return nil, domainerrors.ErrInvalidInput
		}
		selectedIDs[itemID] = struct{}{}
	}
	var createdOrder *models.Order
	err := u.unitOfWork.WithinTransaction(ctx, func(repos interfaces.TransactionRepositories) error {
		cart, err := repos.Cart.GetCartByUserID(ctx, int64(userID))
		if err != nil {
			return err
		}
		cartItems, err := repos.Cart.GetCartItemsByIDs(ctx, int64(cart.ID), cartItemIDs)
		if err != nil {
			return err
		}
		if len(cartItems) != len(cartItemIDs) {
			return domainerrors.ErrInvalidInput
		}
		for _, item := range cartItems {
			if _, exists := selectedIDs[item.ID]; !exists {
				return domainerrors.ErrInvalidInput
			}
		}
		address, err := repos.Address.FindByID(ctx, userID, addressID)
		if err != nil {
			return err
		}
		var subtotal float64
		orderItems := make([]models.OrderItem, 0, len(cartItemIDs))
		variantQuantities := make(map[uint]uint)
		for _, cartItem := range cartItems {
			if cartItem.Quantity <= 0 {
				return domainerrors.ErrInvalidQuantity
			}
			variant := cartItem.Variant
			product := variant.Product
			if !variant.IsActive ||
				!product.IsActive ||
				!product.IsListed ||
				product.DeletedAt != nil ||
				!product.Brand.IsActive ||
				!product.Category.IsActive ||
				product.Brand.DeletedAt != nil ||
				product.Category.DeletedAt != nil {
				return domainerrors.ErrProductNotAvailable
			}
			if variant.SellingPrice <= 0 {
				return domainerrors.ErrInvalidInput
			}
			if variant.Stock < cartItem.Quantity {
				return domainerrors.ErrInsufficientStock
			}
			unitPrice := roundMoney(variant.SellingPrice)
			itemSubtotal := roundMoney(unitPrice * float64(cartItem.Quantity))
			imageURL := ""
			for _, image := range variant.Images {
				if image.IsPrimary {
					imageURL = image.ImageURL
					break
				}
			}
			if imageURL == "" && len(variant.Images) > 0 {
				imageURL = variant.Images[0].ImageURL
			}
			if imageURL == "" {
				for _, image := range product.Images {
					if image.IsPrimary {
						imageURL = image.ImageURL
						break
					}
				}
			}
			if imageURL == "" && len(product.Images) > 0 {
				imageURL = product.Images[0].ImageURL
			}
			variantInfo, err := json.Marshal(map[string]any{
				"sku":   variant.SKU,
				"size":  variant.Size,
				"color": variant.Color,
			})
			if err != nil {
				return err
			}
			orderItems = append(orderItems, models.OrderItem{
				VariantID:   variant.ID,
				ProductName: product.Name,
				ImageURL:    imageURL,
				VariantInfo: variantInfo,
				UnitPrice:   unitPrice,
				Quantity:    cartItem.Quantity,
				Subtotal:    itemSubtotal,
			})
			subtotal = roundMoney(itemSubtotal + subtotal)
			variantQuantities[variant.ID] += uint(cartItem.Quantity)
		}
		if subtotal <= 0 {
			return domainerrors.ErrInvalidInput
		}
		var (
			coupon         *models.Coupon
			discountAmount float64
		)
		if couponCode != nil {
			coupon, err := repos.Coupon.GetByCodeForUpdate(ctx, *couponCode)
			if err != nil {
				return err
			}
			nowForCoupon := time.Now()
			if !coupon.IsActive ||
				nowForCoupon.Before(coupon.StartAt) ||
				!nowForCoupon.Before(coupon.ExpiresAt) {
				return domainerrors.ErrInvalidCoupon
			}
			if subtotal < coupon.MinimumOrderAmount {
				return domainerrors.ErrInvalidCoupon
			}
			if coupon.UsageLimit != nil {
				count, err := repos.CouponUsage.CountByCouponID(
					ctx,
					coupon.ID,
				)
				if err != nil {
					return err
				}

				if count >= int64(*coupon.UsageLimit) {
					return domainerrors.ErrInvalidCoupon
				}
			}
			if coupon.UsageLimitPerUser != nil {
				count, err := repos.CouponUsage.CountByCouponAndUser(
					ctx,
					coupon.ID,
					userID,
				)
				if err != nil {
					return err
				}
				if count >= int64(*coupon.UsageLimitPerUser) {
					return domainerrors.ErrInvalidInput
				}
			}
			switch strings.ToLower(coupon.DiscountType) {
			case "percentage":
				discountAmount = roundMoney(
					subtotal * coupon.DiscountValue / 100,
				)
			case "fixed":
				discountAmount = roundMoney(coupon.DiscountValue)
			default:
				return domainerrors.ErrInvalidInput
			}
			if coupon.MaximumDiscount != nil &&
				discountAmount > *coupon.MaximumDiscount {
				discountAmount = *coupon.MaximumDiscount
			}
			if discountAmount > subtotal {
				discountAmount = subtotal
			}
			discountAmount = roundMoney(discountAmount)
			if discountAmount < 0 {
				return domainerrors.ErrInvalidInput
			}
		}
		now := time.Now()
		var paymentDeadline *time.Time
		orderStatus := "pending"
		if paymentMethod == "cod" {
			orderStatus = "confirmed"
		}
		if paymentMethod == "online" {
			deadline := now.Add(10 * time.Minute)
			paymentDeadline = &deadline
		}
		orderID := uuid.New()
		shippingAmount := 0.0
		totalAmount := roundMoney(
			subtotal - discountAmount + shippingAmount,
		)
		newOrder := &models.Order{
			ID:              orderID,
			UserID:          userID,
			OrderNumber:     "ORD-" + strings.ToUpper(uuid.NewString()),
			Subtotal:        subtotal,
			DiscountAmount:  discountAmount,
			ShippingAmount:  shippingAmount,
			TotalAmount:     totalAmount,
			OrderStatus:     orderStatus,
			PaymentStatus:   "pending",
			PaymentDeadline: paymentDeadline,
			PlacedAt:        now,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := repos.Order.Create(ctx, newOrder); err != nil {
			return err
		}
		history := &models.OrderStatusHistory{
			OrderID:   orderID,
			Status:    orderStatus,
			CreatedAt: now,
		}
		if err := repos.OrderStatusHistory.Create(ctx, history); err != nil {
			return err
		}
		for i := range orderItems {
			orderItems[i].OrderID = orderID
			orderItems[i].CreatedAt = now
			orderItems[i].UpdatedAt = now
		}
		if err := repos.OrderItem.CreateMany(ctx, orderItems); err != nil {
			return err
		}
		orderAddresses := &models.OrderAddress{
			OrderID:      orderID,
			Name:         address.Name,
			Phone:        address.Phone,
			AddressLine1: address.AddressLine1,
			City:         address.City,
			State:        address.State,
			PostalCode:   address.PostalCode,
			Country:      address.Country,
			CreatedAt:    now,
		}
		if address.AddressLine2 != "" {
			line2 := address.AddressLine2
			orderAddresses.AddressLine2 = &line2
		}
		if err := repos.OrderAddress.Create(ctx, orderAddresses); err != nil {
			return err
		}
		payment := &models.Payment{
			OrderID:       orderID,
			PaymentMethod: paymentMethod,
			PaymentStatus: "pending",
			Amount:        newOrder.TotalAmount,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := repos.Payment.Create(ctx, payment); err != nil {
			return err
		}
		if coupon != nil {
			usage := &models.CouponUsage{
				CouponID:       coupon.ID,
				UserID:         userID,
				OrderID:        orderID,
				DiscountAmount: discountAmount,
				UsedAt:         now,
			}

			if err := repos.CouponUsage.Create(ctx, usage); err != nil {
				return err
			}
		}
		for variantID, quantity := range variantQuantities {
			if err := repos.ProductVariant.DecreaseStock(ctx, int64(variantID), int(quantity)); err != nil {
				return err
			}
		}
		if err := repos.Cart.RemoveCartItems(ctx, int64(cart.ID), cartItemIDs); err != nil {
			return err
		}
		eventData, err := json.Marshal(map[string]any{
			"order_id":       newOrder.ID.String(),
			"order_number":   newOrder.OrderNumber,
			"payment_method": paymentMethod,
			"total_amount":   newOrder.TotalAmount,
		})
		if err != nil {
			return err
		}
		entityType := "order"
		entityID := newOrder.ID.String()
		rawEventData := json.RawMessage(eventData)

		notification := &models.Notification{
			ID:         uuid.New(),
			Type:       "order_placed",
			Title:      "New order placed",
			Message:    "Order " + newOrder.OrderNumber + " has been placed.",
			Priority:   "medium",
			EntityType: &entityType,
			EntityID:   &entityID,
			EventData:  &rawEventData,
			CreatedAt:  now,
		}
		adminIDs, err := repos.User.ListAdminIDs(ctx)
		if err != nil {
			return err
		}
		if err := repos.Notification.Create(ctx, notification); err != nil {
			return err
		}
		// Assign the notification to all admins.
		recipients := make([]models.NotificationRecipient, 0, len(adminIDs))
		for _, adminID := range adminIDs {
			recipients = append(recipients, models.NotificationRecipient{
				NotificationID: notification.ID,
				AdminID:        adminID,
			})
		}
		if err := repos.Notification.CreateRecipients(ctx, recipients); err != nil {
			return err
		}

		createdOrder = newOrder
		return nil
	})
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to place order",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}
	u.log.InfoContext(
		ctx,
		"order placed successfully",
		"user_id", userID,
		"order_id", createdOrder.ID,
		"order_number", createdOrder.OrderNumber,
		"total_amount", createdOrder.TotalAmount,
	)
	return createdOrder, nil
}
func (u *OrderUseCase) CancelOrder(ctx context.Context, userID uint, orderID uuid.UUID) error {
	if userID == 0 {
		return domainerrors.ErrInvalidUserID
	}
	if orderID == uuid.Nil {
		return domainerrors.ErrInvalidOrderID
	}
	now := time.Now()
	err := u.unitOfWork.WithinTransaction(ctx, func(repos interfaces.TransactionRepositories) error {
		order, err := repos.Order.LockByID(ctx, orderID)
		if err != nil {
			return err
		}
		if order.UserID != userID {
			return domainerrors.ErrRecordNotFound
		}
		if order.OrderStatus != "pending" &&
			order.OrderStatus != "confirmed" {
			return domainerrors.ErrInvalidOrderTransition
		}
		items, err := repos.OrderItem.ListByOrderID(ctx, order.ID)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return domainerrors.ErrInvalidInput
		}
		for _, item := range items {
			if item.Quantity <= 0 {
				return domainerrors.ErrInvalidQuantity
			}
			if err := repos.ProductVariant.IncreaseStock(ctx, int64(item.VariantID), item.Quantity); err != nil {
				return err
			}
		}
		order.OrderStatus = "cancelled"
		order.UpdatedAt = now
		if err := repos.Order.Update(ctx, order); err != nil {
			return err
		}
		note := "Order cancelled by customer"
		history := &models.OrderStatusHistory{
			OrderID:   order.ID,
			Status:    "cancelled",
			Note:      &note,
			CreatedAt: now,
		}
		if err := repos.OrderStatusHistory.Create(ctx, history); err != nil {
			return err
		}
		if err := repos.CouponUsage.CancelByOrderID(ctx, order.ID); err != nil {
			return err
		}
		eventData, err := json.Marshal(map[string]any{
			"order_id":     order.ID.String(),
			"order_number": order.OrderNumber,
			"status":       "cancelled",
		})
		if err != nil {
			return err
		}
		entityType := "order"
		entityID := order.ID.String()
		rawEventData := json.RawMessage(eventData)
		notification := &models.Notification{
			ID:         uuid.New(),
			Type:       "order_cancelled",
			Title:      "Order cancelled",
			Message:    "Order " + order.OrderNumber + " has been cancelled by the customer.",
			Priority:   "medium",
			EntityType: &entityType,
			EntityID:   &entityID,
			EventData:  &rawEventData,
			CreatedAt:  now,
		}
		if err := repos.Notification.Create(ctx, notification); err != nil {
			return err
		}
		adminIDs, err := repos.User.ListAdminIDs(ctx)
		if err != nil {
			return err
		}
		recipients := make([]models.NotificationRecipient, 0, len(adminIDs))
		for _, adminID := range adminIDs {
			recipients = append(recipients, models.NotificationRecipient{
				NotificationID: notification.ID,
				AdminID:        adminID,
			})
		}
		if len(recipients) > 0 {
			if err := repos.Notification.CreateRecipients(ctx, recipients); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to cancel order",
			"user_id", userID,
			"order_id", orderID,
			"error", err,
		)
		return err
	}
	u.log.InfoContext(
		ctx,
		"order cancelled successfully",
		"user_id", userID,
		"order_id", orderID,
	)
	return nil
}
func (u *OrderUseCase) UpdateOrderAddress(ctx context.Context, userID uint, orderID uuid.UUID, input models.OrderAddress) error {
	if userID == 0 {
		return domainerrors.ErrInvalidUserID
	}

	if orderID == uuid.Nil {
		return domainerrors.ErrInvalidOrderID
	}
	err := u.unitOfWork.WithinTransaction(
		ctx,
		func(repos interfaces.TransactionRepositories) error {
			order, err := repos.Order.LockByID(ctx, orderID)
			if err != nil {
				return err
			}
			if order.UserID != userID {
				return domainerrors.ErrRecordNotFound
			}
			if order.OrderStatus != "pending" &&
				order.OrderStatus != "confirmed" {
				return domainerrors.ErrInvalidOrderTransition
			}
			orderAddress, err := repos.OrderAddress.GetByOrderID(
				ctx,
				order.ID,
			)
			if err != nil {
				return err
			}
			orderAddress.Name = input.Name
			orderAddress.Phone = input.Phone
			orderAddress.AddressLine1 = input.AddressLine1
			orderAddress.AddressLine2 = input.AddressLine2
			orderAddress.City = input.City
			orderAddress.State = input.State
			orderAddress.PostalCode = input.PostalCode
			// orderAddress. = now

			if err := repos.OrderAddress.Update(ctx, orderAddress); err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to update order address",
			"user_id", userID,
			"order_id", orderID,
			"error", err,
		)
		return err
	}

	u.log.InfoContext(
		ctx,
		"order address updated successfully",
		"user_id", userID,
		"order_id", orderID,
	)

	return nil
}
func (u *OrderUseCase) GenerateInvoice(ctx context.Context, userID uint, orderID uuid.UUID) ([]byte, error) {
	if userID == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	if orderID == uuid.Nil {
		return nil, domainerrors.ErrInvalidInput
	}
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("retrieve order for invoice: %w", err)
	}
	if order == nil {
		return nil, domainerrors.ErrRecordNotFound
	}
	if order.UserID != userID {
		return nil, domainerrors.ErrRecordNotFound
	}
	if len(order.Items) == 0 {
		return nil, fmt.Errorf("cannot generate invoice: order has no items")
	}
	if order.Address == nil {
		return nil, fmt.Errorf("cannot generate invoice: order address is missing")
	}
	pdfBytes, err := u.invoiceGenerator.Generate(order)
	if err != nil {
		return nil, fmt.Errorf("generate order invoice: %w", err)
	}
	if len(pdfBytes) == 0 {
		return nil, fmt.Errorf("invoice generator returned an empty PDF")
	}
	return pdfBytes, nil
}
func (u *OrderUseCase) UpdateCODPaymentStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	if orderID == uuid.Nil {
		return domainerrors.ErrInvalidOrderID
	}
	if status != "paid" {
		return domainerrors.ErrInvalidInput
	}
	return u.unitOfWork.WithinTransaction(
		ctx,
		func(repos interfaces.TransactionRepositories) error {
			order, err := repos.Order.LockByID(ctx, orderID)
			if err != nil {
				return err
			}
			if order.PaymentStatus == "paid" {
				return domainerrors.ErrInvalidInput
			}
			payments, err := repos.Payment.GetByOrderID(ctx, orderID)
			if err != nil {
				return err
			}
			var codPayment *models.Payment
			for i := range payments {
				if strings.EqualFold(
					payments[i].PaymentMethod,
					"cod",
				) {
					if codPayment != nil {
						return domainerrors.ErrInvalidInput
					}
					codPayment = &payments[i]
				}
			}
			if codPayment == nil {
				return domainerrors.ErrInvalidInput
			}
			if codPayment.PaymentStatus != "pending" {
				return domainerrors.ErrInvalidInput
			}
			now := time.Now()
			codPayment.PaymentStatus = "paid"
			codPayment.PaidAt = &now
			codPayment.UpdatedAt = now
			if err := repos.Payment.Update(ctx, codPayment); err != nil {
				return err
			}
			order.PaymentStatus = "paid"
			order.UpdatedAt = now
			if err := repos.Order.Update(ctx, order); err != nil {
				return err
			}
			u.log.InfoContext(
				ctx,
				"COD payment marked as paid",
				"order_id", orderID,
				"payment_id", codPayment.ID,
			)
			return nil
		},
	)
}
func roundMoney(amount float64) float64 {
	return math.Round(amount*100) / 100
}

var _ interfaces.OrderUseCase = (*OrderUseCase)(nil)
