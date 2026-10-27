package postgres

import (
	"context"
	"errors"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReturnRequestRepository struct {
	db *gorm.DB
}

func NewReturnRequestRepository(db *gorm.DB) *ReturnRequestRepository {
	return &ReturnRequestRepository{
		db: db,
	}
}

func (r *ReturnRequestRepository) Create(ctx context.Context, request *models.ReturnRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}
func (r *ReturnRequestRepository) GetByID(ctx context.Context, id uint) (*models.ReturnRequest, error) {
	var request models.ReturnRequest
	err := r.db.WithContext(ctx).
		Preload("OrderItem").
		Preload("RequestedVariant").
		Where("id = ?", id).
		First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &request, nil
}
func (r *ReturnRequestRepository) GetByIDAndUserID(ctx context.Context, id uint, userID uint) (*models.ReturnRequest, error) {
	var request models.ReturnRequest
	err := r.db.WithContext(ctx).
		Joins("JOIN orders ON orders.id = return_requests.order_id").
		Where("return_requests.id = ?", id).
		Where("orders.user_id = ?", userID).
		Preload("OrderItem").
		Preload("RequestedVariant").
		First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &request, nil
}
func (r *ReturnRequestRepository) ListByUserID(ctx context.Context, userID uint, limit int, page int) ([]models.ReturnRequest, int64, error) {
	var requests []models.ReturnRequest
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.ReturnRequest{}).
		Joins("JOIN orders ON orders.id = return_requests.order_id").
		Where("orders.user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.
		Preload("OrderItem").
		Preload("RequestedVariant").
		Order("return_requests.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}
func (r *ReturnRequestRepository) List(ctx context.Context, orderID *uuid.UUID, status *string, limit int, page int) ([]models.ReturnRequest, int64, error) {
	var requests []models.ReturnRequest
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.ReturnRequest{})
	if orderID != nil {
		query = query.Where("order_id = ?", *orderID)
	}
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.
		Preload("OrderItem").
		Preload("RequestedVariant").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}
func (r *ReturnRequestRepository) Update(ctx context.Context, request *models.ReturnRequest) error {
	result := r.db.WithContext(ctx).
		Model(&models.ReturnRequest{}).
		Where("id = ?", request.ID).
		Select(
			"Status",
			"AdminNote",
			"UpdatedAt",
		).
		Updates(request)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *ReturnRequestRepository) HasActiveRequest(ctx context.Context, orderItemID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ReturnRequest{}).
		Where("order_item_id = ?", orderItemID).
		Where("status NOT IN ?", []string{
			"rejected",
			"closed",
		}).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

var _ interfaces.ReturnRequestRepository = (*ReturnRequestRepository)(nil)
