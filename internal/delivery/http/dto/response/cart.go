package response

import "ecommerce/internal/domain/models"

type CartResponse struct {
	ID                  uint               `json:"id"`
	Items               []CartItemResponse `json:"items"`
	ItemCount           int                `json:"item_count"`
	TotalQuantity       int                `json:"total_quantity"`
	AvailableSubtotal   float64            `json:"available_subtotal"`
	HasUnavailableItems bool               `json:"has_unavailable_items"`
}

type CartItemResponse struct {
	ID           uint   `json:"id"`
	ProductID    uint   `json:"product_id"`
	ProductName  string `json:"product_name"`
	ProductSlug  string `json:"product_slug"`
	CategoryName string `json:"category_name"`
	BrandName    string `json:"brand_name"`

	VariantID uint    `json:"variant_id"`
	SKU       string  `json:"sku"`
	Size      *string `json:"size,omitempty"`
	Color     *string `json:"color,omitempty"`

	Quantity       int     `json:"quantity"`
	UnitPrice      float64 `json:"unit_price"`
	LineTotal      float64 `json:"line_total"`
	AvailableStock int     `json:"available_stock"`

	IsAvailable        bool    `json:"is_available"`
	AvailabilityReason string  `json:"availability_reason,omitempty"`
	ImageURL           *string `json:"image_url,omitempty"`
}

func getCartItemImage(item models.CartItem) *string {
	for _, image := range item.Variant.Images {
		if image.IsPrimary {
			url := image.ImageURL
			return &url
		}
	}
	if len(item.Variant.Images) > 0 {
		url := item.Variant.Images[0].ImageURL
		return &url
	}
	for _, image := range item.Variant.Product.Images {
		if image.VariantID == nil && image.IsPrimary {
			url := image.ImageURL
			return &url
		}
	}

	return nil
}
func ToCartItemResponse(item models.CartItem) CartItemResponse {
	variant := item.Variant
	product := variant.Product

	isAvailable := true
	availabilityReason := ""
	if product.ID == 0 ||
		!product.IsActive ||
		!product.IsListed ||
		product.DeletedAt != nil ||
		!product.Category.IsActive ||
		!product.Brand.IsActive ||
		product.Category.DeletedAt != nil ||
		product.Brand.DeletedAt != nil {

		isAvailable = false
		availabilityReason = "product_unavailable"

	} else if !variant.IsActive {

		isAvailable = false
		availabilityReason = "variant_inactive"

	} else if variant.Stock <= 0 {

		isAvailable = false
		availabilityReason = "out_of_stock"

	} else if item.Quantity > variant.Stock {

		isAvailable = false
		availabilityReason = "insufficient_stock"
	}
	imageURL := getCartItemImage(item)
	productName := product.Name
	if product.ID == 0 {
		productName = "Unavailable product"
	}
	return CartItemResponse{
		ID:           item.ID,
		ProductID:    variant.ProductID,
		ProductName:  productName,
		ProductSlug:  product.Slug,
		CategoryName: product.Category.Name,
		BrandName:    product.Brand.Name,

		VariantID: variant.ID,
		SKU:       variant.SKU,
		Size:      variant.Size,
		Color:     variant.Color,

		Quantity:       item.Quantity,
		UnitPrice:      variant.SellingPrice,
		LineTotal:      variant.SellingPrice * float64(item.Quantity),
		AvailableStock: variant.Stock,

		IsAvailable:        isAvailable,
		AvailabilityReason: availabilityReason,
		ImageURL:           imageURL,
	}
}
func ToCartResponse(cart *models.Cart) CartResponse {
	response := CartResponse{
		ID:    cart.ID,
		Items: make([]CartItemResponse, 0, len(cart.Items)),
	}
	for _, item := range cart.Items {
		itemResponse := ToCartItemResponse(item)
		response.Items = append(response.Items, itemResponse)
		response.ItemCount++
		response.TotalQuantity += item.Quantity
		if itemResponse.IsAvailable {
			response.AvailableSubtotal += itemResponse.LineTotal
		} else {
			response.HasUnavailableItems = true
		}
	}
	return response
}
