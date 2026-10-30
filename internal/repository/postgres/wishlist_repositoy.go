package postgres

import (
	"context"
	"errors"

	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/usecase/interfaces"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type wishlistRepository struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) *wishlistRepository {
	return &wishlistRepository{
		db: db,
	}
}
func (r *wishlistRepository) CreateWishlist(ctx context.Context, wishlist *models.Wishlist) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"},
				{Name: "name"},
			},
			DoNothing: true}).
		Create(wishlist).Error
}
func (r *wishlistRepository) GetWishlistsByUserID(ctx context.Context, userID uint) ([]models.Wishlist, error) {
	var wishlists []models.Wishlist
	err := r.db.WithContext(ctx).
		Model(&models.Wishlist{}).
		Select(`
		wishlist.*,
		COUNT(wishlists_items.id) AS item_count
		`).
		Joins(`LEFT JOIN wishlist_items
		ON wishlist_items.wishlist_id=wishlist.id
		`).
		Where("wishlist.user_id = ?", userID).
		Group("wishlist.id").
		Order("wishlist.created_at DESC").
		Find(&wishlists).Error
	if err != nil {
		return nil, err
	}
	return wishlists, nil
}
func (r *wishlistRepository) UpdateWishlist(ctx context.Context, wishlistID uint, userID uint, name string) error {
	result := r.db.WithContext(ctx).
		Model(&models.Wishlist{}).
		Where("id = ? AND user_id = ?", wishlistID, userID).
		Update("name", name)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrCheckConstraintViolated) {
			return domainerrors.ErrDuplicateRecord
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *wishlistRepository) DeleteWishlist(ctx context.Context, wishlistID uint, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", wishlistID, userID).
		Delete(&models.Wishlist{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *wishlistRepository) GetWishlistByID(ctx context.Context, userID uint, wishlistID uint) (*models.Wishlist, error) {
	var wishlist models.Wishlist
	result := r.db.WithContext(ctx).
		Model(&models.Wishlist{}).
		Where("id = ? AND user_id = ?", wishlistID, userID).
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Category").
		Preload("Items.Product.Brand").
		Preload("Items.Product.Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_primary = ?", true)
		}).
		First(&wishlist).Error
	if result != nil {
		if errors.Is(result, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, result
	}
	return &wishlist, nil
}
func (r *wishlistRepository) GetWishlistByName(ctx context.Context, userID uint, wishlistName string) (*models.Wishlist, error) {
	var wishlist models.Wishlist
	result := r.db.WithContext(ctx).
		Model(&models.Wishlist{}).
		Where("name = ? AND user_id = ?", wishlistName, userID).
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Category").
		Preload("Items.Product.Brand").
		Preload("Items.Product.Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_primary = ?", true)
		}).
		First(&wishlist).Error
	if result != nil {
		if errors.Is(result, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, result
	}
	return &wishlist, nil
}
func (r *wishlistRepository) RemoveWishlistItem(ctx context.Context, wishlistID uint, userID uint, itemID uint) error {
	err := r.db.WithContext(ctx).
		Where(`
	wishlist_id = ?
	AND id = ?
	AND EXISTS(
		SELECT 1
		FROM wishlist
		WHERE wishlist.id = wishlist_items.id 
		AND wishlist.user_id = ?
	)	
	`, wishlistID, itemID, userID).
		Delete(&models.WishlistItem{})
	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return domainerrors.ErrRecordNotFound
	}
	return nil
}
func (r *wishlistRepository) ClearWishlist(ctx context.Context, wishlistID uint, userID uint) error {
	return r.db.WithContext(ctx).
		Where(`
			wishlist_id = ?
			AND EXISTS(
			SELECT 1 
			FROM wishlist
			WHERE wishlist.id = wishlist_items.wishlist_id 
			AND wishlist.user_id = ?
			)
		`, wishlistID, userID).
		Delete(&models.WishlistItem{}).Error
}
func (r *wishlistRepository) AddProductToWishlist(ctx context.Context, wishlistname string, userID uint, productID uint) error {
	const q = `
		WITH p as (
			SELECT pr.id 
			FROM products pr 
			JOIN categories c ON c.id = pr.category_id 
			JOIN brands b ON b.id = pr.brand_id
			WHERE pr.id = ?
				AND pr.is_active = true 
				AND pr.is_listed = true
				AND pr.deleted_at IS NULL
				AND c.is_active = true
				AND c.deleted_at IS NULL
				AND b.is_active = true
				AND b.deleted_at IS NULL
		),
		w as (
			INSERT INTO wishlist (user_id,name)
			SELECT ?,? FROM p
			ON CONFLICT (user_id,name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		),
		i as (
			INSERT INTO wishlist_items (wishlist_id, product_id)
			SELECT w.id, p.id FROM w CROSS JOIN p
			ON CONFLICT (wishlist_id, product_id) DO NOTHING
			RETURNING 1
		)
		SELECT
			(SELECT COUNT(*) FROM p) AS active,
			(SELECT COUNT(*) FROM i) AS added
		`
	var out struct {
		Active int
		Added  int
	}
	err := r.db.WithContext(ctx).Raw(q, productID, userID, wishlistname).Scan(&out).Error
	if err != nil {
		return err
	}
	if out.Active == 0 {
		return domainerrors.ErrProductNotAvailable
	}
	if out.Added == 0 {
		return domainerrors.ErrWishlistItemAlreadyExists
	}
	return nil
}
func (r *wishlistRepository) MergeWishlistItems(ctx context.Context, wishlistName string, userID uint, productIDs []uint) (uint, uint, error) {
	const q = `
		WITH p AS(
			SELECT pr.id
			FROM products pr
			JOIN categories c on c.id = pr.category_id
			JOIN brands b on b.id = pr.brand_id
			WHERE pr.id IN ?
				AND pr.is_active = true
				AND pr.is_listed = true
				AND pr.deleted_at IS NULL
				AND c.is_active = true
				AND c.deleted_at IS NULL
				AND b.is_active = true
				AND b.deleted_at IS NULL
		),
		w AS(
			INSERT INTO wishlist (user_id,name)
			SELECT ?,? WHERE EXISTS (SELECT 1 FROM p)
			ON CONFLICT (user_id,name) DO UPDATE SET name = EXCLUDED.name
			RETURNING ID
		)
		INSERT INTO wishlist_items (wishlist_id,product_id)
		SELECT w.id,p.id FROM w CROSS JOIN p
		ON CONFLICT (wishlist_id,product_id) DO NOTHING
	`
	res := r.db.WithContext(ctx).Exec(q, productIDs, userID, wishlistName)
	if res.Error != nil {
		return 0, 0, res.Error
	}
	accepted := uint(res.RowsAffected)
	skipped := uint(len(productIDs) - int(accepted))
	return accepted, skipped, nil
}

var _ interfaces.WishlistRepository = (*wishlistRepository)(nil)
