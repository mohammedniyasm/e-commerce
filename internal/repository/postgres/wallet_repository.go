package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *walletRepository {
	return &walletRepository{
		db: db,
	}
}
func (r *walletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return domainerrors.ErrUserNotFound
		}
		return err
	}
	return nil
}
func (r *walletRepository) GetByUserID(ctx context.Context, userID uint) (*models.Wallet, error) {
	db := r.db.WithContext(ctx)
	var wallet models.Wallet
	err := db.First(&wallet, userID).Error
	if err == nil {
		return &wallet, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err := r.Create(ctx, &models.Wallet{ID: userID}); err != nil {
		return nil, err
	}
	if err := db.First(&wallet, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &wallet, nil
}
func (r *walletRepository) GetByUserIDForUpdate(ctx context.Context, userID uint) (*models.Wallet, error) {
	db := r.db.WithContext(ctx)
	var wallet models.Wallet
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&wallet, userID).Error
	if err == nil {
		return &wallet, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err := r.Create(ctx, &models.Wallet{ID: userID}); err != nil {
		return nil, err
	}
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&wallet, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &wallet, nil
}
func (r *walletRepository) Update(ctx context.Context, wallet *models.Wallet) error {
	result := r.db.WithContext(ctx).
		Model(&models.Wallet{}).
		Where("id = ?", wallet.ID).
		Select("Balance", "UpdatedAt").
		Updates(wallet)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *walletRepository) CreateTransaction(ctx context.Context, transaction *models.WalletTransaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}
func (r *walletRepository) GetTransactionByID(ctx context.Context, transactionID uint, walletID uint) (*models.WalletTransaction, error) {
	var transaction models.WalletTransaction
	err := r.db.WithContext(ctx).
		Where("id = ? AND wallet_id = ?", transactionID, walletID).
		First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &transaction, nil
}
func (r *walletRepository) ListTransactions(ctx context.Context, walletID uint, limit int, page int, sort *string) ([]models.WalletTransaction, int64, error) {
	var transactions []models.WalletTransaction
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.WalletTransaction{}).
		Where("wallet_id = ?", walletID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if sort != nil {
		switch *sort {
		case "old":
			query = query.Order("created_at ASC")
		case "new":
			query = query.Order("created_at DESC")
		default:
			query = query.Order("created_at DESC")
		}
	} else {
		query = query.Order("created_at DESC")
	}
	query = query.Order("id DESC")
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	err := query.
		Select(
			"id",
			"transaction_type",
			"source",
			"amount",
			"reference_id",
			"description",
			"created_at",
		).
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}
	return transactions, total, nil
}

var _ interfaces.WalletRepository = (*walletRepository)(nil)
