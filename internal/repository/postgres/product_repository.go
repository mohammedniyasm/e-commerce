package postgres

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	err := r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ? AND deleted_at is null", product.ID).Updates(map[string]interface{}{
		"name":              product.Name,
		"slug":              product.Slug,
		"short_description": product.ShortDescription,
		"description":       product.Description,
		"category_id":       product.CategoryID,
		"brand_id":          product.BrandID,
	}).Error
	return err
}
func (r *ProductRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &product, nil
}
func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &product, nil
}
func (r *ProductRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{
		"is_active":  false,
		"is_listed":  false,
		"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}).Error
}
func (r *ProductRepository) List(ctx context.Context, search string, page int, limit int, isActive *bool, isListed *bool, categoryID *int64, brandID *int64) ([]models.Product, int64, error) {

	var products []models.Product
	var total int64

	countQuery := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("deleted_at IS NULL")

	if search != "" {
		searchPattern := "%" + search + "%"

		countQuery = countQuery.Where(
			"name ILIKE ? OR slug ILIKE ?",
			searchPattern,
			searchPattern,
		)
	}

	if isActive != nil {
		countQuery = countQuery.Where(
			"is_active = ?",
			*isActive,
		)
	}

	if isListed != nil {
		countQuery = countQuery.Where(
			"is_listed = ?",
			*isListed,
		)
	}

	if categoryID != nil {
		countQuery = countQuery.Where(
			"category_id = ?",
			*categoryID,
		)
	}

	if brandID != nil {
		countQuery = countQuery.Where(
			"brand_id = ?",
			*brandID,
		)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("deleted_at IS NULL")

	if search != "" {
		searchPattern := "%" + search + "%"

		dataQuery = dataQuery.Where(
			"name ILIKE ? OR slug ILIKE ?",
			searchPattern,
			searchPattern,
		)
	}

	if isActive != nil {
		dataQuery = dataQuery.Where(
			"is_active = ?",
			*isActive,
		)
	}

	if isListed != nil {
		dataQuery = dataQuery.Where(
			"is_listed = ?",
			*isListed,
		)
	}

	if categoryID != nil {
		dataQuery = dataQuery.Where(
			"category_id = ?",
			*categoryID,
		)
	}

	if brandID != nil {
		dataQuery = dataQuery.Where(
			"brand_id = ?",
			*brandID,
		)
	}

	offset := (page - 1) * limit

	err := dataQuery.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&products).
		Error

	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
func (r *ProductRepository) ListDeleted(ctx context.Context, search string, page int, limit int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64
	query := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("deleted_at IS NOT NULL")
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"name ILIKE ? OR slug ILIKE ?",
			searchPattern,
			searchPattern,
		)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	err := query.
		Order("deleted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&products).
		Error
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}
func (r *ProductRepository) GetDeletedByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Where(
			"id = ? AND deleted_at IS NOT NULL",
			id,
		).
		First(&product).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &product, nil
}
func (r *ProductRepository) Restore(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(
			"id = ? AND deleted_at IS NOT NULL",
			id,
		).
		Updates(map[string]interface{}{
			"is_active":  true,
			"is_listed":  true,
			"deleted_at": nil,
		}).
		Error
}
func (r *ProductRepository) ToggleActive(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(
			"id = ? AND deleted_at IS NULL",
			id,
		).
		Update(
			"is_active",
			gorm.Expr("NOT is_active"),
		).
		Error
}
func (r *ProductRepository) ToggleListed(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(
			"id = ? AND deleted_at IS NULL",
			id,
		).
		Update(
			"is_listed",
			gorm.Expr("NOT is_listed"),
		).
		Error
}
func (r *ProductRepository) CountActiveByCategoryID(ctx context.Context, categoryID int64) (int64, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(
			"category_id = ? AND is_active = true AND deleted_at IS NULL",
			categoryID,
		).
		Count(&count).
		Error

	return count, err
}
func (r *ProductRepository) CountActiveByBrandID(ctx context.Context, brandID int64) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(
			"brand_id = ? AND is_active = true AND deleted_at IS NULL",
			brandID,
		).
		Count(&count).
		Error

	return count, err
}
func (r *ProductRepository) ListStoreProducts(
	ctx context.Context,
	search string,
	categoryID *int64,
	brandID *int64,
	minPrice *float64,
	maxPrice *float64,
	sort string,
	page int,
	limit int,
) ([]models.Product, int64, error) {

	baseQuery := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Joins(
			"JOIN product_variants pv ON pv.product_id = products.id AND pv.is_active = ?",
			true,
		).
		Where(
			"products.is_active = ? AND products.is_listed = ? AND products.deleted_at IS NULL",
			true,
			true,
		)

	if search != "" {
		searchPattern := "%" + search + "%"
		baseQuery = baseQuery.Where(
			"products.name ILIKE ? OR products.slug ILIKE ?",
			searchPattern,
			searchPattern,
		)
	}

	// Category filter.
	if categoryID != nil {
		baseQuery = baseQuery.Where(
			"products.category_id = ?",
			*categoryID,
		)
	}

	// Brand filter.
	if brandID != nil {
		baseQuery = baseQuery.Where(
			"products.brand_id = ?",
			*brandID,
		)
	}
	if minPrice != nil || maxPrice != nil {

		priceQuery := `
			EXISTS (
				SELECT 1
				FROM product_variants pv_filter
				WHERE pv_filter.product_id = products.id
				AND pv_filter.is_active = true
				
		`

		args := make([]interface{}, 0, 2)

		if minPrice != nil {
			priceQuery += " AND pv_filter.selling_price >= ?"
			args = append(args, *minPrice)
		}

		if maxPrice != nil {
			priceQuery += " AND pv_filter.selling_price <= ?"
			args = append(args, *maxPrice)
		}

		priceQuery += ")"

		baseQuery = baseQuery.Where(priceQuery, args...)
	}

	// Count unique products.
	var total int64

	countQuery := baseQuery.Session(&gorm.Session{})

	if err := countQuery.
		Select("COUNT(DISTINCT products.id)").
		Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Product query.
	dataQuery := baseQuery.
		Select("products.*").
		Group("products.id")

	// Sorting.
	switch sort {

	case "price_asc":
		dataQuery = dataQuery.Order(
			"MIN(pv.selling_price) ASC, products.created_at DESC",
		)

	case "price_desc":
		dataQuery = dataQuery.Order(
			"MIN(pv.selling_price) DESC, products.created_at DESC",
		)

	case "name_asc":
		dataQuery = dataQuery.Order(
			"products.name ASC, products.created_at DESC",
		)

	case "name_desc":
		dataQuery = dataQuery.Order(
			"products.name DESC, products.created_at DESC",
		)

	default:
		dataQuery = dataQuery.Order(
			"products.created_at DESC",
		)
	}

	offset := (page - 1) * limit

	var products []models.Product

	err := dataQuery.
		Offset(offset).
		Limit(limit).
		Preload("Category").
		Preload("Brand").
		Preload(
			"Variants",
			"is_active = ?",
			true,
		).
		Preload(
			"Images",
			"is_primary = ?",
			true,
		).
		Find(&products).
		Error

	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
func (r *ProductRepository) GetStoreProductByID(ctx context.Context, id int64) (*models.Product, error) {

	if id <= 0 {
		return nil, domainerrors.ErrInvalidProductID
	}

	return r.getStoreProduct(
		ctx,
		"products.id = ?",
		id,
	)
}
func (r *ProductRepository) GetStoreProductBySlug(ctx context.Context, slug string) (*models.Product, error) {

	slug = strings.TrimSpace(slug)

	if slug == "" {
		return nil, domainerrors.ErrInvalidProductSlug
	}

	return r.getStoreProduct(
		ctx,
		"products.slug = ?",
		slug,
	)
}
func (r *ProductRepository) getStoreProduct(ctx context.Context, condition string, value interface{}) (*models.Product, error) {

	var product models.Product

	err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where(condition, value).
		Where(
			"products.is_active = ? AND products.is_listed = ? AND products.deleted_at IS NULL",
			true,
			true,
		).
		Where(`
			EXISTS (
				SELECT 1
				FROM product_variants pv
				WHERE pv.product_id = products.id
				AND pv.is_active = true
			)
		`).
		Preload("Category").
		Preload("Brand").
		Preload(
			"Variants",
			"is_active = ?",
			true,
		).
		Preload(
			"Images",
			func(db *gorm.DB) *gorm.DB {
				return db.Order("display_order ASC, id ASC")
			},
		).
		Preload(
			"Variants.Images",
			func(db *gorm.DB) *gorm.DB {
				return db.Order("display_order ASC, id ASC")
			},
		).
		First(&product).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}

		return nil, err
	}

	return &product, nil
}
func (r *ProductRepository) GetRelatedProducts(ctx context.Context,productID int64,limit int) ([]models.Product, error) {
	var source models.Product
	err := r.db.WithContext(ctx).
		Select("id", "category_id", "brand_id").
		Where("id = ?", productID).
		Where("is_active = ?", true).
		Where("is_listed = ?", true).
		First(&source).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRecordNotFound
		}

		return nil, err
	}

	// Prevent an unbounded query.
	if limit <= 0 {
		limit = 8
	}

	if limit > 20 {
		limit = 20
	}

	var products []models.Product

	err = r.db.WithContext(ctx).
		Model(&models.Product{}).

		// Do not recommend the product currently being viewed.
		Where("products.id <> ?", productID).

		// Recommend only active and listed products.
		Where("products.is_active = ?", true).
		Where("products.is_listed = ?", true).

		// Match either the category or the brand.
		Where(
			"(products.category_id = ? OR products.brand_id = ?)",
			source.CategoryID,
			source.BrandID,
		).

		// A recommended product must have at least one active variant.
		Where(`
            EXISTS (
                SELECT 1
                FROM product_variants AS pv
                WHERE pv.product_id = products.id
                  AND pv.is_active = ?
            )
        `, true).

		// Load the data needed to display recommended products.
		Preload("Category").
		Preload("Brand").
		Preload("Variants", "is_active = ?", true).
		Preload(
			"Images",
			"variant_id IS NULL AND is_primary = ?",
			true,
		).
		Order(clause.Expr{
			SQL: `
                CASE
                    WHEN products.category_id = ? THEN 0
                    ELSE 1
                END,
                CASE
                    WHEN products.brand_id = ? THEN 0
                    ELSE 1
                END
            `,
			Vars: []interface{}{
				source.CategoryID,
				source.BrandID,
			},
			WithoutParentheses: true,
		}).
		Order("products.created_at DESC").
		Limit(limit).
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	return products, nil
}
