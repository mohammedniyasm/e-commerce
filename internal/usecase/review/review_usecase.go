package review

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
)

const (
	minRating    int16 = 1
	maxRating    int16 = 5
	maxPageLimit       = 100
)

type ReviewUseCase struct {
	reviewRepo         interfaces.ReviewRepository
	orderItemRepo      interfaces.OrderItemRepository
	orderRepo          interfaces.OrderRepository
	shipmentRepo       interfaces.ShipmentRepository
	productVariantRepo interfaces.ProductVariantRepository
	log                *slog.Logger
}

func NewReviewUseCase(
	reviewRepo interfaces.ReviewRepository,
	orderItemRepo interfaces.OrderItemRepository,
	orderRepo interfaces.OrderRepository,
	shipmentRepo interfaces.ShipmentRepository,
	productVariantRepo interfaces.ProductVariantRepository,
	log *slog.Logger,
) *ReviewUseCase {
	return &ReviewUseCase{
		reviewRepo:         reviewRepo,
		orderItemRepo:      orderItemRepo,
		orderRepo:          orderRepo,
		shipmentRepo:       shipmentRepo,
		productVariantRepo: productVariantRepo,
		log:                log,
	}
}

func validateRating(rating int16) error {
	if rating < minRating || rating > maxRating {
		return domainerrors.ErrInvalidInput
	}
	return nil
}
func normalizeComment(comment *string) *string {
	if comment == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*comment)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
func (u *ReviewUseCase) CreateReview(ctx context.Context, userID uint, orderItemID uint, rating int16, comment *string) (*models.Review, error) {
	if userID == 0 || orderItemID == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	if err := validateRating(rating); err != nil {
		return nil, err
	}
	existingReview, err := u.reviewRepo.GetByOrderItemID(
		ctx,
		orderItemID,
	)
	if err == nil && existingReview != nil {
		return nil, domainerrors.ErrDuplicateRecord
	}
	if err != nil && !errors.Is(err, domainerrors.ErrRecordNotFound) {
		return nil, err
	}
	orderItem, err := u.orderItemRepo.GetByID(ctx, orderItemID)
	if err != nil {
		return nil, err
	}
	order, err := u.orderRepo.GetByID(ctx, orderItem.OrderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, domainerrors.ErrRecordNotFound
	}
	variant, err := u.productVariantRepo.GetByID(
		ctx,
		int64(orderItem.VariantID),
	)
	if err != nil {
		return nil, err
	}
	shipment, err := u.shipmentRepo.GetByOrderID(
		ctx,
		orderItem.OrderID,
	)
	if err != nil {
		return nil, err
	}
	if shipment.Type != models.ShipmentTypeDelivery ||
		shipment.ShippingStatus != "delivered" {
		return nil, domainerrors.ErrInvalidInput
	}
	review := &models.Review{
		ProductID:   variant.ProductID,
		UserID:      userID,
		OrderItemID: orderItemID,
		Rating:      rating,
		Comment:     normalizeComment(comment),
		IsVisible:   true,
	}
	if err := u.reviewRepo.Create(ctx, review); err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to create product review",
			"user_id", userID,
			"order_item_id", orderItemID,
			"error", err,
		)
		return nil, err
	}
	u.log.InfoContext(
		ctx,
		"product review created successfully",
		"review_id", review.ID,
		"user_id", userID,
		"product_id", variant.ProductID,
	)
	return review, nil
}
func (u *ReviewUseCase) GetMyReview(ctx context.Context, userID uint, reviewID uint) (*models.Review, error) {
	if userID == 0 || reviewID == 0 {
		return nil, domainerrors.ErrInvalidInput
	}
	return u.reviewRepo.GetByID(ctx, reviewID, userID)
}
func (u *ReviewUseCase) ListMyReviews(ctx context.Context, userID uint, page int, limit int) ([]models.Review, int64, error) {
	if userID == 0 {
		return nil, 0, domainerrors.ErrInvalidInput
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return u.reviewRepo.ListByUserID(
		ctx,
		userID,
		page,
		limit,
	)
}
func (u *ReviewUseCase) UpdateReview(ctx context.Context, userID uint, reviewID uint, rating int16, comment *string) error {
	if userID == 0 || reviewID == 0 {
		return domainerrors.ErrInvalidInput
	}
	if err := validateRating(rating); err != nil {
		return err
	}
	err := u.reviewRepo.Update(
		ctx,
		reviewID,
		userID,
		rating,
		normalizeComment(comment),
	)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to update product review",
			"review_id", reviewID,
			"user_id", userID,
			"error", err,
		)
		return err
	}
	return nil
}
func (u *ReviewUseCase) DeleteReview(ctx context.Context, userID uint, reviewID uint) error {
	if userID == 0 || reviewID == 0 {
		return domainerrors.ErrInvalidInput
	}
	err := u.reviewRepo.Delete(ctx, reviewID, userID)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to delete product review",
			"review_id", reviewID,
			"user_id", userID,
			"error", err,
		)
		return err
	}
	return nil
}
func (u *ReviewUseCase) ListProductReviews(ctx context.Context, productID uint, page int, limit int) ([]models.Review, int64, error) {
	if productID == 0 {
		return nil, 0, domainerrors.ErrInvalidInput
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return u.reviewRepo.ListByProductID(
		ctx,
		productID,
		page,
		limit,
	)
}
func (u *ReviewUseCase) ListAllReviews(ctx context.Context, page int, limit int) ([]models.Review, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return u.reviewRepo.ListAll(ctx, page, limit)
}
func (u *ReviewUseCase) SetReviewVisibility(ctx context.Context, reviewID uint, isVisible bool) error {
	if reviewID == 0 {
		return domainerrors.ErrInvalidInput
	}
	err := u.reviewRepo.SetVisibility(
		ctx,
		reviewID,
		isVisible,
	)
	if err != nil {
		u.log.ErrorContext(
			ctx,
			"failed to update review visibility",
			"review_id", reviewID,
			"is_visible", isVisible,
			"error", err,
		)
		return err
	}
	return nil
}

var _ interfaces.ReviewUseCase = (*ReviewUseCase)(nil)
