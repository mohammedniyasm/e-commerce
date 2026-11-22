package response

import (
	"encoding/json"
	"testing"

	"ecommerce/internal/domain/models"

	"github.com/google/uuid"
)

func TestToReturnRequestResponseIncludesItemAndExchangeDetails(t *testing.T) {
	orderID := uuid.New()
	size := "M"
	color := "Blue"
	requestedVariantID := uint(42)
	request := &models.ReturnRequest{
		ID:                 7,
		OrderID:            orderID,
		OrderItemID:        9,
		RequestType:        "exchange",
		RequestedVariantID: &requestedVariantID,
		Quantity:           1,
		Reason:             "Need a different size",
		Status:             "requested",
		Order: models.Order{
			ID:          orderID,
			OrderNumber: "ORD-1007",
			OrderStatus: "delivered",
			UserID:      3,
		},
		OrderItem: models.OrderItem{
			ID:          9,
			VariantID:   21,
			ProductName: "Everyday Shirt",
			ImageURL:    "https://example.test/shirt.jpg",
			VariantInfo: json.RawMessage(`{"size":"S","color":"Blue"}`),
			UnitPrice:   1299,
			Quantity:    2,
			Subtotal:    2598,
		},
		RequestedVariant: &models.ProductVariant{
			ID:           requestedVariantID,
			ProductID:    12,
			Size:         &size,
			Color:        &color,
			SKU:          "SHIRT-M-BLUE",
			SellingPrice: 1399,
			Stock:        4,
			Images: []models.ProductImage{
				{
					ProductID: 12,
					VariantID: &requestedVariantID,
					ImageURL:  "https://example.test/shirt-m-blue.jpg",
					IsPrimary: true,
				},
			},
		},
	}

	got := ToReturnRequestResponse(request)

	if got.Order == nil || got.Order.OrderNumber != "ORD-1007" {
		t.Fatalf("expected order summary to be included, got %#v", got.Order)
	}
	if got.OrderItem == nil ||
		got.OrderItem.ProductName != "Everyday Shirt" ||
		got.OrderItem.VariantID != 21 {
		t.Fatalf("expected original order item details to be included, got %#v", got.OrderItem)
	}
	if got.RequestedVariant == nil ||
		got.RequestedVariant.ID != requestedVariantID ||
		got.RequestedVariant.SKU != "SHIRT-M-BLUE" ||
		got.RequestedVariant.ImageURL != "https://example.test/shirt-m-blue.jpg" {
		t.Fatalf("expected requested exchange variant details to be included, got %#v", got.RequestedVariant)
	}
}
