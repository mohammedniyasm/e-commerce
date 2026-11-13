package review

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	domainErrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type reviewFixture struct {
	uc            *ReviewUseCase
	reviewRepo    *mocks.MockReviewRepository
	orderItemRepo *mocks.MockOrderItemRepository
	orderRepo     *mocks.MockOrderRepository
	shipmentRepo  *mocks.MockShipmentRepository
	variantRepo   *mocks.MockProductVariantRepository
}

func newReviewFixture(t *testing.T) *reviewFixture {
	t.Helper()
	reviewRepo := mocks.NewMockReviewRepository(t)
	orderItemRepo := mocks.NewMockOrderItemRepository(t)
	orderRepo := mocks.NewMockOrderRepository(t)
	shipmentRepo := mocks.NewMockShipmentRepository(t)
	variantRepo := mocks.NewMockProductVariantRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &reviewFixture{
		uc:            NewReviewUseCase(reviewRepo, orderItemRepo, orderRepo, shipmentRepo, variantRepo, log),
		reviewRepo:    reviewRepo,
		orderItemRepo: orderItemRepo,
		orderRepo:     orderRepo,
		shipmentRepo:  shipmentRepo,
		variantRepo:   variantRepo,
	}
}

func TestReviewHelpers(t *testing.T) {
	assert.NoError(t, validateRating(1))
	assert.NoError(t, validateRating(5))
	assert.Error(t, validateRating(0))
	assert.Error(t, validateRating(6))

	blank := "   "
	assert.Nil(t, normalizeComment(&blank))
	assert.Nil(t, normalizeComment(nil))
	text := "  Great shirt  "
	normalized := normalizeComment(&text)
	require.NotNil(t, normalized)
	assert.Equal(t, "Great shirt", *normalized)
}

func TestReviewUseCase_CreateReview(t *testing.T) {
	ctx := context.Background()
	orderID := uuid.New()
	comment := "  Great fit  "

	t.Run("validation errors", func(t *testing.T) {
		f := newReviewFixture(t)
		_, err := f.uc.CreateReview(ctx, 0, 10, 5, &comment)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)

		_, err = f.uc.CreateReview(ctx, 1, 0, 5, &comment)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)

		_, err = f.uc.CreateReview(ctx, 1, 10, 0, &comment)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)
	})

	t.Run("ownership and delivery checks", func(t *testing.T) {
		f := newReviewFixture(t)
		item := &models.OrderItem{ID: 10, OrderID: orderID, VariantID: 22, Quantity: 1}

		f.reviewRepo.EXPECT().GetByOrderItemID(ctx, uint(10)).Return(nil, domainErrors.ErrRecordNotFound).Once()
		f.orderItemRepo.EXPECT().GetByID(ctx, uint(10)).Return(item, nil).Once()
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 99, OrderStatus: "delivered"}, nil).Once()
		_, err := f.uc.CreateReview(ctx, 1, 10, 5, &comment)
		assert.Error(t, err)

		f.reviewRepo.EXPECT().GetByOrderItemID(ctx, uint(10)).Return(nil, domainErrors.ErrRecordNotFound).Once()
		f.orderItemRepo.EXPECT().GetByID(ctx, uint(10)).Return(item, nil).Once()
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{ID: orderID, UserID: 1, OrderStatus: "shipped"}, nil).Once()
		f.variantRepo.EXPECT().GetByID(ctx, int64(22)).Return(&models.ProductVariant{ID: 22, ProductID: 77}, nil).Maybe()
		f.shipmentRepo.EXPECT().GetByOrderID(ctx, orderID).Return(&models.Shipment{OrderID: orderID, ShippingStatus: "in_transit"}, nil).Maybe()
		_, err = f.uc.CreateReview(ctx, 1, 10, 5, &comment)
		assert.Error(t, err)
	})

	t.Run("duplicate review check", func(t *testing.T) {
		f := newReviewFixture(t)
		f.reviewRepo.EXPECT().GetByOrderItemID(ctx, uint(10)).Return(&models.Review{ID: 5}, nil).Once()

		_, err := f.uc.CreateReview(ctx, 1, 10, 5, &comment)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		f := newReviewFixture(t)
		now := time.Now()
		item := &models.OrderItem{ID: 10, OrderID: orderID, VariantID: 22, Quantity: 1}

		f.reviewRepo.EXPECT().GetByOrderItemID(ctx, uint(10)).Return(nil, domainErrors.ErrRecordNotFound).Once()
		f.orderItemRepo.EXPECT().GetByID(ctx, uint(10)).Return(item, nil).Once()
		f.orderRepo.EXPECT().GetByID(ctx, orderID).Return(&models.Order{
			ID: orderID, UserID: 1, OrderStatus: "delivered", PaymentStatus: "paid",
		}, nil).Once()
		f.shipmentRepo.EXPECT().GetByOrderID(ctx, orderID).Return(&models.Shipment{
			OrderID: orderID, Type: models.ShipmentTypeDelivery, ShippingStatus: "delivered", DeliveredAt: &now,
		}, nil).Maybe()
		f.variantRepo.EXPECT().GetByID(ctx, int64(22)).Return(&models.ProductVariant{ID: 22, ProductID: 77}, nil).Maybe()
		f.reviewRepo.EXPECT().Create(ctx, mock.MatchedBy(func(r *models.Review) bool {
			return r.UserID == 1 && r.OrderItemID == 10 && r.Rating == 4 && r.Comment != nil && *r.Comment == "Great fit"
		})).Return(nil).Once()

		got, err := f.uc.CreateReview(ctx, 1, 10, 4, &comment)
		require.NoError(t, err)
		assert.Equal(t, int16(4), got.Rating)
	})
}

func TestReviewUseCase_QueriesMutationsAndVisibility(t *testing.T) {
	ctx := context.Background()

	t.Run("GetMyReview and ListMyReviews", func(t *testing.T) {
		f := newReviewFixture(t)
		_, err := f.uc.GetMyReview(ctx, 0, 5)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)

		_, err = f.uc.GetMyReview(ctx, 1, 0)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)

		expected := &models.Review{ID: 5, UserID: 1}
		f.reviewRepo.EXPECT().GetByID(ctx, uint(5), uint(1)).Return(expected, nil).Once()
		got, err := f.uc.GetMyReview(ctx, 1, 5)
		require.NoError(t, err)
		assert.Equal(t, expected, got)

		_, _, err = f.uc.ListMyReviews(ctx, 0, 1, 10)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)

		list := []models.Review{*expected}
		f.reviewRepo.EXPECT().ListByUserID(ctx, uint(1), 1, 10).Return(list, int64(1), nil).Once()
		gotList, total, err := f.uc.ListMyReviews(ctx, 1, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, list, gotList)
		assert.Equal(t, int64(1), total)
	})

	t.Run("UpdateReview and DeleteReview", func(t *testing.T) {
		f := newReviewFixture(t)
		text := " Updated comment "
		assert.ErrorIs(t, f.uc.UpdateReview(ctx, 0, 5, 4, &text), domainErrors.ErrInvalidInput)
		assert.ErrorIs(t, f.uc.UpdateReview(ctx, 1, 0, 4, &text), domainErrors.ErrInvalidInput)
		assert.ErrorIs(t, f.uc.UpdateReview(ctx, 1, 5, 9, &text), domainErrors.ErrInvalidInput)

		f.reviewRepo.EXPECT().Update(ctx, uint(5), uint(1), int16(4), mock.MatchedBy(func(c *string) bool {
			return c != nil && *c == "Updated comment"
		})).Return(nil).Once()
		require.NoError(t, f.uc.UpdateReview(ctx, 1, 5, 4, &text))

		assert.ErrorIs(t, f.uc.DeleteReview(ctx, 0, 5), domainErrors.ErrInvalidInput)
		assert.ErrorIs(t, f.uc.DeleteReview(ctx, 1, 0), domainErrors.ErrInvalidInput)

		delErr := errors.New("delete failed")
		f.reviewRepo.EXPECT().Delete(ctx, uint(5), uint(1)).Return(delErr).Once()
		assert.ErrorIs(t, f.uc.DeleteReview(ctx, 1, 5), delErr)
	})

	t.Run("ListProductReviews ListAllReviews and SetReviewVisibility", func(t *testing.T) {
		f := newReviewFixture(t)
		_, _, err := f.uc.ListProductReviews(ctx, 0, 1, 10)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)

		list := []models.Review{{ID: 9}}
		f.reviewRepo.EXPECT().ListByProductID(ctx, uint(12), 1, 10).Return(list, int64(1), nil).Once()
		got, total, err := f.uc.ListProductReviews(ctx, 12, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, list, got)
		assert.Equal(t, int64(1), total)

		f.reviewRepo.EXPECT().ListAll(ctx, 1, 10).Return(list, int64(1), nil).Once()
		got, total, err = f.uc.ListAllReviews(ctx, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, list, got)
		assert.Equal(t, int64(1), total)

		assert.ErrorIs(t, f.uc.SetReviewVisibility(ctx, 0, true), domainErrors.ErrInvalidInput)
		f.reviewRepo.EXPECT().SetVisibility(ctx, uint(9), false).Return(nil).Once()
		require.NoError(t, f.uc.SetReviewVisibility(ctx, 9, false))
	})
}
